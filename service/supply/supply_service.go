package supply

import (
	"compound/core"
	"context"

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
		marketService:  marketService,
	}
}

// 赎回
func (s *supplyService) Redeem(ctx context.Context, redeemTokens decimal.Decimal, market *core.Market) (string, error) {
	return "", nil
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
	return "", nil
}

//抵押
func (s *supplyService) Pledge(ctx context.Context, pledgedTokens decimal.Decimal, market *core.Market) (string, error) {
	return "", nil
}

//撤销抵押
func (s *supplyService) Unpledge(ctx context.Context, unpledgedTokens decimal.Decimal, userID string, market *core.Market) error {
	return nil
}
