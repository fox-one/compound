package oracle

import (
	"time"

	"github.com/shopspring/decimal"
)

// GetUnderlyingPrice get underlying price with asset id
func GetUnderlyingPrice(assetID string, t time.Time) (decimal.Decimal, error) {
	//TODO: default decimal.Zero
	return decimal.Zero, nil
}
