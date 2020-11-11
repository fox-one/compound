package supply

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"errors"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

type supplyService struct {
	config         *core.Config
	mainWallet     *mixin.Client
	blockWallet    *mixin.Client
	db             *db.DB
	supplyStore    core.ISupplyStore
	marketStore    core.IMarketStore
	accountService core.IAccountService
	priceService   core.IPriceOracleService
	blockService   core.IBlockService
	walletService  core.IWalletService
}

// New new supply service
func New(cfg *core.Config,
	mainWallet *mixin.Client,
	db *db.DB,
	supplyStore core.ISupplyStore,
	marketStore core.IMarketStore,
	accountService core.IAccountService,
	priceService core.IPriceOracleService,
	blockService core.IBlockService,
	walletService core.IWalletService) core.ISupplyService {
	return &supplyService{
		config:         cfg,
		mainWallet:     mainWallet,
		db:             db,
		supplyStore:    supplyStore,
		marketStore:    marketStore,
		accountService: accountService,
		priceService:   priceService,
		blockService:   blockService,
		walletService:  walletService,
	}
}

// 赎回
func (s *supplyService) Redeem(ctx context.Context, redeemTokens decimal.Decimal, userID string, market *core.Market) (string, error) {
	if !s.RedeemAllowed(ctx, redeemTokens, userID, market) {
		return "", errors.New("redeem not allowed")
	}

	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceRedeem

	str, e := action.Format()
	if e != nil {
		return "", e
	}

	return s.walletService.PaySchemaURL(redeemTokens, market.CTokenAssetID, s.mainWallet.ClientID, id.GenTraceID(), str)
}

func (s *supplyService) RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, userID string, market *core.Market) bool {
	supply, e := s.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return false
	}

	remainTokens := supply.CTokens.Sub(supply.CollateTokens)
	// check ctokens
	if redeemTokens.GreaterThan(remainTokens) {
		return false
	}

	amount := supply.Principal.Mul(redeemTokens).Div(supply.CTokens)

	// check market cash
	marketCash, e := s.mainWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return false
	}

	if amount.GreaterThan(marketCash.Balance) {
		return false
	}

	return true
}

// return max redeem ctokens
func (s *supplyService) MaxRedeem(ctx context.Context, userID string, market *core.Market) (decimal.Decimal, error) {
	supply, e := s.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return decimal.Zero, e
	}

	remainTokens := supply.CTokens.Sub(supply.CollateTokens)
	remainAmount := supply.Principal.Mul(remainTokens).Div(supply.CTokens)

	// check market cash
	marketCash, e := s.mainWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return decimal.Zero, e
	}

	if remainAmount.GreaterThan(marketCash.Balance) {
		remainAmount = marketCash.Balance
	}

	remainTokens = supply.CTokens.Mul(remainAmount).Div(supply.Principal)

	return remainTokens, nil
}

// 存入
func (s *supplyService) Supply(ctx context.Context, amount decimal.Decimal, market *core.Market) (string, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return "", errors.New("invalid amount")
	}

	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceSupply

	str, e := action.Format()
	if e != nil {
		return "", e
	}
	return s.walletService.PaySchemaURL(amount, market.AssetID, s.mainWallet.ClientID, id.GenTraceID(), str)
}

//抵押
func (s *supplyService) Pledge(ctx context.Context, pledgedTokens decimal.Decimal, userID string, market *core.Market) (string, error) {
	supply, e := s.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return "", e
	}

	remainTokens := supply.CTokens.Sub(supply.CollateTokens)
	if pledgedTokens.GreaterThan(remainTokens) {
		return "", errors.New("insufficient remain tokens")
	}

	memo := make(core.Action)
	memo[core.ActionKeyService] = core.ActionServicePledge

	memoStr, e := memo.Format()
	if e != nil {
		return "", e
	}

	return s.walletService.PaySchemaURL(pledgedTokens, market.CTokenAssetID, s.mainWallet.ClientID, id.GenTraceID(), memoStr)
}

//撤销抵押
func (s *supplyService) Unpledge(ctx context.Context, unpledgedTokens decimal.Decimal, userID string, market *core.Market) error {
	supply, e := s.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return e
	}

	if unpledgedTokens.GreaterThanOrEqual(supply.CollateTokens) {
		return errors.New("invalid unpledge tokens")
	}

	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		return e
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return e
	}

	price, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
	if e != nil {
		return e
	}

	unpledgedTokenLiquidity := unpledgedTokens.Mul(supply.Principal).Div(supply.CTokens).Mul(market.CollateralFactor).Mul(price)
	if unpledgedTokenLiquidity.GreaterThanOrEqual(liquidity) {
		return errors.New("insufficient liquidity")
	}

	trace := id.UUIDFromString(fmt.Sprintf("unpledge-%s-%s-%d", userID, market.Symbol, curBlock))
	input := mixin.TransferInput{
		AssetID:    s.config.App.BlockAssetID,
		OpponentID: s.mainWallet.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    trace,
	}

	if !s.walletService.VerifyPayment(ctx, &input) {
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceUnpledge
		memo[core.ActionKeyCToken] = unpledgedTokens.String()
		memo[core.ActionKeySymbol] = market.Symbol
		memo[core.ActionKeyUser] = userID

		memoStr, e := memo.Format()
		if e != nil {
			return e
		}

		input.Memo = memoStr

		_, e = s.blockWallet.Transfer(ctx, &input, s.config.BlockWallet.Pin)
		if e != nil {
			return e
		}
	}

	return nil
}

//当前最大可抵押
func (s *supplyService) MaxPledge(ctx context.Context, userID string, market *core.Market) (decimal.Decimal, error) {
	supply, e := s.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return decimal.Zero, e
	}

	return supply.CTokens.Sub(supply.CollateTokens), nil
}
