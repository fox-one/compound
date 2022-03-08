package account

import (
	"compound/core"
	"compound/pkg/compound"
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

type accountService struct {
	marketStore core.IMarketStore
	supplyStore core.ISupplyStore
	borrowStore core.IBorrowStore
}

// New new account service
func New(
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
) core.IAccountService {
	return &accountService{
		marketStore: marketStore,
		supplyStore: supplyStore,
		borrowStore: borrowStore,
	}
}

// CalculateAccountLiquidity calculate account liquidity
//
// 	supplyValue = supply.collaterals * market.exchange_rate * market.collateral_factor * market.price
// 	borrowValue = borrow.Balance()
// 	liquidity = total_supply_values - total_borrow_values
func (s *accountService) CalculateAccountLiquidity(ctx context.Context, userID string, newMarkets ...*core.Market) (decimal.Decimal, error) {
	supplies, e := s.supplyStore.FindByUser(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}
	supplyValue := decimal.Zero
	for _, supply := range supplies {
		market, e := s.findMarketByCtokenAssetID(ctx, newMarkets, supply.CTokenAssetID)
		if e != nil {
			market, e = s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
			if e != nil {
				return decimal.Zero, e
			}
		}

		if market.ID == 0 {
			return decimal.Zero, errors.New("no market")
		}

		price := market.Price
		exchangeRate := market.ExchangeRate
		value := supply.Collaterals.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
		supplyValue = supplyValue.Add(value)
	}

	borrows, e := s.borrowStore.FindByUser(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}

	borrowValue := decimal.Zero

	for _, borrow := range borrows {
		market, e := s.findMarketByAssetID(ctx, newMarkets, borrow.AssetID)
		if e != nil {
			market, e = s.marketStore.Find(ctx, borrow.AssetID)
			if e != nil {
				return decimal.Zero, e
			}
		}

		if market.ID == 0 {
			return decimal.Zero, errors.New("no market")
		}

		price := market.Price

		borrowBalance := compound.BorrowBalance(ctx, borrow, market)
		value := borrowBalance.Mul(price)
		borrowValue = borrowValue.Add(value)
	}

	liquidity := supplyValue.Sub(borrowValue)

	return liquidity, nil
}

// SeizeTokenAllowed
//
// check account liquidity
func (s *accountService) SeizeTokenAllowed(ctx context.Context, supply *core.Supply, borrow *core.Borrow, liquidity decimal.Decimal) bool {
	if supply.UserID != borrow.UserID {
		return false
	}

	// check liquidity
	if liquidity.GreaterThanOrEqual(decimal.Zero) {
		return false
	}

	return true
}

func (s *accountService) findMarketByAssetID(ctx context.Context, src []*core.Market, assetID string) (*core.Market, error) {
	if src == nil {
		return nil, errors.New("no market found")
	}

	for _, m := range src {
		if m.AssetID == assetID {
			return m, nil
		}
	}

	return nil, errors.New("no market found")
}

func (s *accountService) findMarketByCtokenAssetID(ctx context.Context, src []*core.Market, ctokenAssetID string) (*core.Market, error) {
	if src == nil {
		return nil, errors.New("no market found")
	}

	for _, m := range src {
		if m.CTokenAssetID == ctokenAssetID {
			return m, nil
		}
	}

	return nil, errors.New("no market found")
}
