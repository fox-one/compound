package supply

import (
	"compound/core"
	"context"

	"github.com/shopspring/decimal"
)

type supplyService struct {
	marketService core.IMarketService
}

// New new supply service
func New(
	marketService core.IMarketService) core.ISupplyService {
	return &supplyService{
		marketService: marketService,
	}
}

func (s *supplyService) RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, market *core.Market) bool {
	amount := redeemTokens.Mul(market.ExchangeRate)
	supplies := market.TotalCash.Sub(market.Reserves)
	return supplies.GreaterThan(amount)
}
