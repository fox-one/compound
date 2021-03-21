package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Account user account model
// includes liquidity, supplies and borrows
type Account struct {
	UserID    string          `json:"user_id"`
	Liquidity decimal.Decimal `json:"liquidity"`
	Supplies  []*Supply       `json:"supplies"`
	Borrows   []*Borrow       `json:"borrows"`
}

// IAccountService account service interface
type IAccountService interface {
	// calculate account liquidity
	CalculateAccountLiquidity(ctx context.Context, userID string, blockNum int64) (decimal.Decimal, error)
	MaxSeize(ctx context.Context, supply *Supply, borrow *Borrow) (decimal.Decimal, error)
	SeizeTokenAllowed(ctx context.Context, supply *Supply, borrow *Borrow, time time.Time) bool
	SeizeToken(ctx context.Context, supply *Supply, borrow *Borrow, repayAmount decimal.Decimal) (string, error)
}
