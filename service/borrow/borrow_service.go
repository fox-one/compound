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
	blockService   core.IBlockService
	priceService   core.IPriceOracleService
	accountService core.IAccountService
}

// New new borrow service
func New(
	blockService core.IBlockService,
	priceService core.IPriceOracleService,
	accountService core.IAccountService) core.IBorrowService {
	return &borrowService{
		blockService:   blockService,
		priceService:   priceService,
		accountService: accountService,
	}
}

// BorrowAllowed check borrow capacity, check account liquidity
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

// Deprecated
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
