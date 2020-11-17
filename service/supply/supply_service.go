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
	db             *db.DB
	mainWallet     *core.Wallet
	blockWallet    *core.Wallet
	supplyStore    core.ISupplyStore
	marketStore    core.IMarketStore
	accountService core.IAccountService
	priceService   core.IPriceOracleService
	blockService   core.IBlockService
	walletService  core.IWalletService
	marketService  core.IMarketService
}

// New new supply service
func New(cfg *core.Config,
	db *db.DB,
	mainWallet *core.Wallet,
	blockWallet *core.Wallet,
	supplyStore core.ISupplyStore,
	marketStore core.IMarketStore,
	accountService core.IAccountService,
	priceService core.IPriceOracleService,
	blockService core.IBlockService,
	walletService core.IWalletService,
	marketService core.IMarketService) core.ISupplyService {
	return &supplyService{
		config:         cfg,
		db:             db,
		mainWallet:     mainWallet,
		blockWallet:    blockWallet,
		supplyStore:    supplyStore,
		marketStore:    marketStore,
		accountService: accountService,
		priceService:   priceService,
		blockService:   blockService,
		walletService:  walletService,
		marketService:  marketService,
	}
}

// 赎回
func (s *supplyService) Redeem(ctx context.Context, redeemTokens decimal.Decimal, market *core.Market) (string, error) {
	if !s.RedeemAllowed(ctx, redeemTokens, market) {
		return "", errors.New("redeem not allowed")
	}

	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceRedeem

	str, e := action.Format()
	if e != nil {
		return "", e
	}

	return s.walletService.PaySchemaURL(redeemTokens, market.CTokenAssetID, s.mainWallet.Client.ClientID, id.GenTraceID(), str)
}

func (s *supplyService) RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, market *core.Market) bool {
	exchangeRate, e := s.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return false
	}
	amount := redeemTokens.Mul(exchangeRate)
	supplies := market.TotalCash.Sub(market.Reserves)
	if amount.GreaterThan(supplies) {
		return false
	}

	return true
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
	return s.walletService.PaySchemaURL(amount, market.AssetID, s.mainWallet.Client.ClientID, id.GenTraceID(), str)
}

//抵押
func (s *supplyService) Pledge(ctx context.Context, pledgedTokens decimal.Decimal, market *core.Market) (string, error) {
	memo := make(core.Action)
	memo[core.ActionKeyService] = core.ActionServicePledge

	memoStr, e := memo.Format()
	if e != nil {
		return "", e
	}

	return s.walletService.PaySchemaURL(pledgedTokens, market.CTokenAssetID, s.mainWallet.Client.ClientID, id.GenTraceID(), memoStr)
}

//撤销抵押
func (s *supplyService) Unpledge(ctx context.Context, unpledgedTokens decimal.Decimal, userID string, market *core.Market) error {
	supply, e := s.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		return e
	}

	if unpledgedTokens.GreaterThanOrEqual(supply.Collaterals) {
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

	price, e := s.priceService.GetUnderlyingPrice(ctx, market.Symbol, curBlock)
	if e != nil {
		return e
	}

	exchangeRate, e := s.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return e
	}

	unpledgedTokenLiquidity := unpledgedTokens.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
	if unpledgedTokenLiquidity.GreaterThanOrEqual(liquidity) {
		return errors.New("insufficient liquidity")
	}

	trace := id.UUIDFromString(fmt.Sprintf("unpledge-%s-%s-%d", userID, market.Symbol, curBlock))
	input := mixin.TransferInput{
		AssetID:    s.config.App.BlockAssetID,
		OpponentID: s.mainWallet.Client.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    trace,
	}

	if !s.walletService.VerifyPayment(ctx, &input) {
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceUnpledge
		memo[core.ActionKeyUser] = userID
		memo[core.ActionKeySymbol] = market.Symbol
		memo[core.ActionKeyCToken] = unpledgedTokens.Truncate(8).String()

		memoStr, e := memo.Format()
		if e != nil {
			return e
		}

		input.Memo = memoStr
		_, e = s.blockWallet.Client.Transfer(ctx, &input, s.blockWallet.Pin)
		if e != nil {
			return e
		}
	}

	return nil
}

// TODO
func (s *supplyService) MaxUnpledge(ctx context.Context, userID string, market *core.Market) (decimal.Decimal, error) {

	return decimal.Zero, nil
}
