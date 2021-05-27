package number

import (
	"math"

	"github.com/shopspring/decimal"
)

func Decimal(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		return decimal.Zero
	}
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

func Floor(d decimal.Decimal, precision int32) decimal.Decimal {
	return d.Shift(precision).Floor().Shift(-precision)
}
