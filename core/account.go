package core

import (
	"context"

	"github.com/shopspring/decimal"
)

// IAccountService account service interface
type IAccountService interface {
	// calculate account liquidity
	CalculateAccountLiquidity(ctx context.Context, userID string, newMarkets ...*Market) (decimal.Decimal, error)
	SeizeTokenAllowed(ctx context.Context, supply *Supply, borrow *Borrow, liquidity decimal.Decimal) bool
}
