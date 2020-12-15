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
	exchangeRate, e := s.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return false
	}

	amount := redeemTokens.Mul(exchangeRate)
	supplies := market.TotalCash.Sub(market.Reserves)
	if amount.GreaterThan(supplies) {
		return false
	}

	return true
}
