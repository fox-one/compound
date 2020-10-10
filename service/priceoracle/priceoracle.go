package priceoracle

import (
	"time"

	"github.com/shopspring/decimal"
)

// GetUnderlyingPrice get underlying price with asset id
func GetUnderlyingPrice(assetID string, t time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

// SetUnderlyingPrice set underlying price
func SetUnderlyingPrice(assetID string, price decimal.Decimal) error {
	return nil
}
