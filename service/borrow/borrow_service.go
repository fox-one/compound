package borrow

import (
	"compound/core"
	"context"
	"errors"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

type borrowService struct {
	config         *core.Config
	mainWallet     *core.Wallet
	blockWallet    *core.Wallet
	marketStore    core.IMarketStore
	borrowStore    core.IBorrowStore
	blockService   core.IBlockService
	priceService   core.IPriceOracleService
	accountService core.IAccountService
	marketService  core.IMarketService
}

// New new borrow service
func New(cfg *core.Config,
	mainWallet *core.Wallet,
	blockWallet *core.Wallet,
	marketStore core.IMarketStore,
	borrowStore core.IBorrowStore,
	blockService core.IBlockService,
	priceService core.IPriceOracleService,
	accountService core.IAccountService,
	marketService core.IMarketService) core.IBorrowService {
	return &borrowService{
		config:         cfg,
		mainWallet:     mainWallet,
		blockWallet:    blockWallet,
		marketStore:    marketStore,
		borrowStore:    borrowStore,
		blockService:   blockService,
		priceService:   priceService,
		accountService: accountService,
		marketService:  marketService,
	}
}

func (s *borrowService) Repay(ctx context.Context, amount decimal.Decimal, borrow *core.Borrow) (string, error) {
	return "", nil
}

func (s *borrowService) Borrow(ctx context.Context, borrowAmount decimal.Decimal, userID string, market *core.Market) error {
	return nil
}

func (s *borrowService) BorrowAllowed(ctx context.Context, borrowAmount decimal.Decimal, userID string, market *core.Market, time time.Time) bool {
	log := logger.FromContext(ctx)

	if borrowAmount.LessThanOrEqual(decimal.Zero) {
		log.Errorln("invalid borrow amount")
		return false
	}

	// check borrow cap
	supplies := market.TotalCash.Sub(market.Reserves)
	if supplies.LessThan(market.BorrowCap) {
		log.Errorln("insufficient market cash")
		return false
	}

	if borrowAmount.GreaterThan(supplies.Sub(market.BorrowCap)) {
		log.Errorln("insufficient market cash")
		return false
	}

	blockNum, e := s.blockService.GetBlock(ctx, time)
	if e != nil {
		log.Errorln(e)
		return false
	}

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID, blockNum)
	if e != nil {
		log.Errorln(e)
		return false
	}

	price, e := s.priceService.GetCurrentUnderlyingPrice(ctx, market)
	if e != nil {
		log.Errorln(e)
		return false
	}

	borrowValue := borrowAmount.Mul(price)
	if borrowValue.GreaterThan(liquidity) {
		log.Errorln("insufficient liquidity")
		return false
	}

	return true
}

//Deprecated
func (s *borrowService) MaxBorrow(ctx context.Context, userID string, market *core.Market) (decimal.Decimal, error) {
	// check borrow cap
	supplies := market.TotalCash.Sub(market.Reserves)
	if supplies.LessThan(market.BorrowCap) {
		return decimal.Zero, errors.New("insufficient market cash")
	}

	blockNum, e := s.blockService.GetBlock(ctx, time.Now())
	if e != nil {
		return decimal.Zero, e
	}

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID, blockNum)
	if e != nil {
		return decimal.Zero, e
	}

	price, e := s.priceService.GetCurrentUnderlyingPrice(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	borrowAmount := liquidity.Div(price)

	return borrowAmount, nil
}

func (s *borrowService) BorrowBalance(ctx context.Context, borrow *core.Borrow, market *core.Market) (decimal.Decimal, error) {
	return borrow.Balance(ctx, market)
}
