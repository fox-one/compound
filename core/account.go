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
	CalculateAccountLiquidity(ctx context.Context, userID string) (decimal.Decimal, error)
	// >0, greater than liquidity; else less than liquidity
	CompValueAndLiquidity(ctx context.Context, valueAmount decimal.Decimal, valueSymbol string, liqudity decimal.Decimal) (decimal.Decimal, error)
	HasBorrows(ctx context.Context, userID string) bool
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
