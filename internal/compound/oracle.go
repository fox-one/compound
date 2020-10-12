package compound

import (
	"time"

	"github.com/shopspring/decimal"
)

// IOracle orcale interface
type IOracle interface {
	GetCurrentBlockTime() time.Time
	GetUnderlyingPrice(t time.Time) (decimal.Decimal, error)
}
