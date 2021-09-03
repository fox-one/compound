package core

import (
	"context"

	"github.com/shopspring/decimal"
)

// IAccountService account service interface
type IAccountService interface {
	// calculate account liquidity
	CalculateAccountLiquidity(ctx context.Context, userID string, newMarkets ...*Market) (decimal.Decimal, error)
	MaxSeize(ctx context.Context, supply *Supply, borrow *Borrow) (decimal.Decimal, error)
	SeizeTokenAllowed(ctx context.Context, supply *Supply, borrow *Borrow, liquidity decimal.Decimal) bool
	SeizeToken(ctx context.Context, supply *Supply, borrow *Borrow, repayAmount decimal.Decimal) (string, error)
}
