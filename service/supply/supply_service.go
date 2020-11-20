package supply

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"errors"

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
	return nil
}
