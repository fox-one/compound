package number

import (
	"math"

	"github.com/shopspring/decimal"
)

func Decimal(v string) decimal.Decimal {
	d, _ := decimal.NewFromString(v)
	return d
}

func Sqrt(d decimal.Decimal) decimal.Decimal {
	f, _ := d.Float64()
	f = math.Sqrt(f)
	return decimal.NewFromFloat(f)
}

func Ceil(d decimal.Decimal, precision int32) decimal.Decimal {
	return d.Shift(precision).Ceil().Shift(-precision)
}
