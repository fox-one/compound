package views

import "github.com/shopspring/decimal"

// Account account view
type Account struct {
	Liquidity decimal.Decimal `json:"liquidity"`
}
