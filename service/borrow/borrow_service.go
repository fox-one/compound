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
	accountService core.IAccountService
}

// New new borrow service
func New(
	blockService core.IBlockService,
	accountService core.IAccountService) core.IBorrowService {
	return &borrowService{
		blockService:   blockService,
		accountService: accountService,
	}
}

// BorrowAllowed check borrow capacity, check account liquidity
func (s *borrowService) BorrowAllowed(ctx context.Context, borrowAmount decimal.Decimal, userID string, market *core.Market, time time.Time) bool {
	log := logger.FromContext(ctx)

	if !borrowAmount.IsPositive() {
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

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		log.Errorln(e)
		return false
	}

	price := market.Price
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

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}

	price := market.Price
	borrowAmount := liquidity.Div(price)

	return borrowAmount, nil
}

func (s *borrowService) BorrowBalance(ctx context.Context, borrow *core.Borrow, market *core.Market) (decimal.Decimal, error) {
	return borrow.Balance(ctx, market)
}
