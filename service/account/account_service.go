package account

import (
	"compound/core"
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

type accountService struct {
	marketStore   core.IMarketStore
	supplyStore   core.ISupplyStore
	borrowStore   core.IBorrowStore
	blockService  core.IBlockService
	marketService core.IMarketService
}

// New new account service
func New(
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	blockSrv core.IBlockService,
	marketServie core.IMarketService,
) core.IAccountService {
	return &accountService{
		marketStore:   marketStore,
		supplyStore:   supplyStore,
		borrowStore:   borrowStore,
		blockService:  blockSrv,
		marketService: marketServie,
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

		borrowBalance, e := borrow.Balance(ctx, market)
		if e != nil {
			return decimal.Zero, e
		}
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

func (s *accountService) MaxSeize(ctx context.Context, supply *core.Supply, borrow *core.Borrow) (decimal.Decimal, error) {
	if supply.UserID != borrow.UserID {
		return decimal.Zero, errors.New("different user bettween supply and borrow")
	}

	supplyMarket, e := s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
	if e != nil {
		return decimal.Zero, e
	}

	if supplyMarket.ID == 0 {
		return decimal.Zero, errors.New("no market")
	}

	exchangeRate, e := s.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		return decimal.Zero, e
	}

	maxSeize := supply.Collaterals.Mul(exchangeRate).Mul(supplyMarket.CloseFactor)

	supplyPrice := supplyMarket.Price
	borrowMarket, e := s.marketStore.Find(ctx, borrow.AssetID)
	if e != nil {
		return decimal.Zero, e
	}
	if borrowMarket.ID == 0 {
		return decimal.Zero, errors.New("no market")
	}

	borrowPrice := borrowMarket.Price
	seizePrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
	seizeValue := maxSeize.Mul(seizePrice)
	borrowValue := borrow.Principal.Mul(borrowPrice)
	if seizeValue.GreaterThan(borrowValue) {
		seizeValue = borrowValue
		maxSeize = seizeValue.Div(seizePrice)
	}

	return maxSeize, nil
}

func (s *accountService) SeizeToken(ctx context.Context, supply *core.Supply, borrow *core.Borrow, repayAmount decimal.Decimal) (string, error) {
	panic("implement me")
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
