package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// IAccountService account service interface
type IAccountService interface {
	// calculate account liquidity
	CalculateAccountLiquidity(ctx context.Context, userID string) (decimal.Decimal, error)
	MaxSeize(ctx context.Context, supply *Supply, borrow *Borrow) (decimal.Decimal, error)
	SeizeTokenAllowed(ctx context.Context, supply *Supply, borrow *Borrow, time time.Time) bool
	SeizeToken(ctx context.Context, supply *Supply, borrow *Borrow, repayAmount decimal.Decimal) (string, error)
}
