package core

import (
	"context"

	"github.com/shopspring/decimal"
)

// Account 借贷账户
type Account struct {
	Supplies []*Supply `json:"supplies"`
	Borrows  []*Borrow `json:"borrows"`
}

// IAccountService account service interface
type IAccountService interface {
	// calculate account liquidity by real time
	CalculateAccountLiquidity(ctx context.Context, userID string) (decimal.Decimal, error)
	HasBorrows(ctx context.Context, userID string) (bool, error)
}

type IAccount interface {
	getLiquidity() decimal.Decimal
	LiquidateBorrowAllowed() bool
	Redeem()
	RedeemAllowed() bool
	Supply()
	Borrow()
	BorrowAllowed() bool
	RepayBorrow()
	RpayBorrowAllowed() bool
	SeizedTokenAllowed() bool
	SeizeTokens()
	CalculateSeizeTokens()
}
