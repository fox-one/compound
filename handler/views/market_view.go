package views

import (
	"compound/core"

	"github.com/shopspring/decimal"
)

// Market market view
type Market struct {
	core.Market
	SupplyAPY decimal.Decimal `json:"supply_apy"`
	BorrowAPY decimal.Decimal `json:"borrow_apy"`
	Suppliers int64           `json:"suppliers"`
	Borrowers int64           `json:"borrowers"`
}
