package views

import (
	"compound/core"

	"github.com/shopspring/decimal"
)

// Market market view
type Market struct {
	core.Market
	SupplyAPY       decimal.Decimal `json:"supply_apy"`
	BorrowAPY       decimal.Decimal `json:"borrow_apy"`
	TotalBorrow     decimal.Decimal `json:"total_borrow"`
	TotalSupply     decimal.Decimal `json:"total_supply"`
	Suppliers       int64           `json:"suppliers"`
	Borrowers       int64           `json:"borrowers"`
	ExchangeRate    decimal.Decimal `json:"exchange_rate"`
	UtilizationRate decimal.Decimal `json:"utilization_rate"`
}
