package compound

import (
	"compound/core"
	"context"

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
