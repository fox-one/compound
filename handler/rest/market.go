package rest

import (
	"compound/core"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"

	"github.com/shopspring/decimal"
)

// response all market infos
func allMarketsHandler(marketStr core.IMarketStore, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore, marketSrv core.IMarketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		markets, e := marketStr.All(ctx)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		marketViews := make([]*views.Market, 0)
		for _, m := range markets {
			marketView := getMarketView(ctx, m, supplyStr, borrowStr, marketSrv)
			marketViews = append(marketViews, marketView)
		}

		var response struct {
			Data interface{} `json:"data"`
		}
		response.Data = marketViews
		render.JSON(w, response)
	}
}

func getMarketView(ctx context.Context, market *core.Market, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore, marketSrv core.IMarketService) *views.Market {
	supplyRate, e := marketSrv.CurSupplyRate(ctx, market)
	if e != nil {
		supplyRate = decimal.Zero
	}
	borrowRate, e := marketSrv.CurBorrowRate(ctx, market)
	if e != nil {
		borrowRate = decimal.Zero
	}

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
