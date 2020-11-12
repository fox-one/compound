package views

import (
	"compound/core"

	"github.com/shopspring/decimal"
)

// Borrow supply view
type Borrow struct {
	core.Borrow
	Price decimal.Decimal `json:"price"`
}
