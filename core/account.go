package core

import "github.com/shopspring/decimal"

// Account 借贷账户
type Account struct {
	Orders []*Order `json:"orders"`
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
