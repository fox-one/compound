package number

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestCeil(t *testing.T) {
	data := map[string]string{
		"0.10304":     "0.11",
		"0.100000001": "0.11",
		"0.108":       "0.11",
	}

	for k, v := range data {
		t.Run(k, func(t *testing.T) {
			_k := Decimal(k)
			c := Ceil(Decimal(k), 2)
			t.Log(k, c, _k.Round(2))
			assert.Equal(t, v, c.String(), "should be ceil")
		})
	}
}
