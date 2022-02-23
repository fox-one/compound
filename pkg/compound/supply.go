package compound

import (
	"compound/core"
	"context"

	"github.com/shopspring/decimal"
)

func RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, market *core.Market) bool {
	amount := redeemTokens.Mul(market.ExchangeRate)
	supplies := market.TotalCash.Sub(market.Reserves)
	return supplies.GreaterThan(amount)
}
