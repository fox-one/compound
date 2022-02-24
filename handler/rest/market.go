package rest

import (
	"compound/core"
	"compound/handler/render"
	"compound/handler/views"
	"compound/pkg/compound"
	"context"
	"net/http"

	"github.com/shopspring/decimal"
)

// response all market infos
func allMarketsHandler(marketStr core.IMarketStore, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		markets, e := marketStr.All(ctx)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		marketViews := make([]*views.Market, 0)
		for _, m := range markets {
			marketView := getMarketView(ctx, m, supplyStr, borrowStr)
			marketViews = append(marketViews, marketView)
		}

		var response struct {
			Data interface{} `json:"data"`
		}
		response.Data = marketViews
		render.JSON(w, response)
	}
}

func getMarketView(ctx context.Context, market *core.Market, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore) *views.Market {
	supplyRate := CurSupplyRate(market)
	borrowRate := CurBorrowRate(market)
	countOfSupplies, e := supplyStr.CountOfSuppliers(ctx, market.CTokenAssetID)
	if e != nil {
		countOfSupplies = 0
	}

	countOfBorrows, e := borrowStr.CountOfBorrowers(ctx, market.AssetID)
	if e != nil {
		countOfBorrows = 0
	}

	marketView := views.Market{
		Market:    *market,
		SupplyAPY: supplyRate,
		BorrowAPY: borrowRate,
		Suppliers: countOfSupplies,
		Borrowers: countOfBorrows,
	}

	return &marketView
}

// CurBorrowRate current borrow APY
func CurBorrowRate(market *core.Market) decimal.Decimal {
	borrowRatePerBlock := compound.GetBorrowRatePerBlock(
		compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves),
		market.BaseRate,
		market.Multiplier,
		market.JumpMultiplier,
		market.Kink,
	)
	return borrowRatePerBlock.Mul(compound.BlocksPerYear).Truncate(compound.MaxPricision)
}

// CurSupplyRate current supply APY
func CurSupplyRate(market *core.Market) decimal.Decimal {
	supplyRatePerBlock := compound.GetSupplyRatePerBlock(
		compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves),
		market.BaseRate,
		market.Multiplier,
		market.JumpMultiplier,
		market.Kink,
		market.ReserveFactor,
	)
	return supplyRatePerBlock.Mul(compound.BlocksPerYear).Truncate(compound.MaxPricision)
}
