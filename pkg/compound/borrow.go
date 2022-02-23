package compound

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// Balance caculate borrow balance
// balance = borrow.principal * market.borrow_index / borrow.interest_index
func BorrowBalance(ctx context.Context, b *core.Borrow, market *core.Market) (decimal.Decimal, error) {
	if !market.BorrowIndex.IsPositive() {
		market.BorrowIndex = decimal.New(1, 0)
	}

	if !b.InterestIndex.IsPositive() {
		b.InterestIndex = market.BorrowIndex
	}

	principalTimesIndex := b.Principal.Mul(market.BorrowIndex)
	result := principalTimesIndex.Div(b.InterestIndex).
		Shift(MaxPricision).Ceil().Shift(-MaxPricision)

	return result, nil
}

// BorrowAllowed check borrow capacity, check account liquidity
func BorrowAllowed(ctx context.Context, borrowAmount decimal.Decimal, userID string, market *core.Market, liquidity decimal.Decimal) bool {
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
	price := market.Price
	borrowValue := borrowAmount.Mul(price)
	if borrowValue.GreaterThan(liquidity) {
		log.Errorln("insufficient liquidity")
		return false
	}

	return true
}
