package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Account 借贷账户
type Account struct {
	UserID    string          `json:"user_id"`
	Liquidity decimal.Decimal `json:"liquidity"`
	Supplies  []*Supply       `json:"supplies"`
	Borrows   []*Borrow       `json:"borrows"`
}

// IAccountStore account store interface
type IAccountStore interface {
	SaveLiquidity(ctx context.Context, userID string, curBlock int64, liquidity decimal.Decimal) error
	FindLiquidity(ctx context.Context, userID string, curBlock int64) (decimal.Decimal, error)
}

// IAccountService account service interface
type IAccountService interface {
	// calculate account liquidity by real time
	CalculateAccountLiquidity(ctx context.Context, userID string, blockNum int64) (decimal.Decimal, error)
	MaxSeize(ctx context.Context, supply *Supply, borrow *Borrow) (decimal.Decimal, error)
	SeizeTokenAllowed(ctx context.Context, supply *Supply, borrow *Borrow, repayAmount decimal.Decimal, time time.Time) bool
	SeizeToken(ctx context.Context, supply *Supply, borrow *Borrow, repayAmount decimal.Decimal) (string, error)
	SeizeAllowedAccounts(ctx context.Context) ([]*Account, error)
}
