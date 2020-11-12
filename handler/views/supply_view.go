package views

import (
	"compound/core"

	"github.com/shopspring/decimal"
)

// Supply supply view
type Supply struct {
	core.Supply
	Price decimal.Decimal `json:"price"`
}
