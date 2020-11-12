package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
)

func allMarketsHandler(ctx context.Context, marketStr core.IMarketStore, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore, marketSrv core.IMarketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		render.JSON(w, marketViews)
	}
}

func marketHandler(ctx context.Context, marketStr core.IMarketStore, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore, marketSrv core.IMarketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Symbol string `json:"symbol"`
		}
		if err := param.Binding(r, &params); err != nil {
			render.BadRequest(w, err)
			return
		}
		symbol := strings.ToUpper(params.Symbol)
		market, e := marketStr.FindBySymbol(ctx, symbol)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		marketView := getMarketView(ctx, market, supplyStr, borrowStr, marketSrv)

		render.JSON(w, marketView)
	}
}

func getMarketView(ctx context.Context, market *core.Market, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore, marketSrv core.IMarketService) *views.Market {
	exchangeRate, e := marketSrv.CurExchangeRate(ctx, market)
	if e != nil {
		exchangeRate = decimal.Zero
	}

	utilizationRate, e := marketSrv.CurUtilizationRate(ctx, market)
	if e != nil {
		utilizationRate = decimal.Zero
	}

	supplyRate, e := marketSrv.CurSupplyRate(ctx, market)
	if e != nil {
		supplyRate = decimal.Zero
	}
	borrowRate, e := marketSrv.CurBorrowRate(ctx, market)
	if e != nil {
		borrowRate = decimal.Zero
	}

	countOfSupplies, e := supplyStr.CountOfSupplies(ctx, market.Symbol)
	if e != nil {
		countOfSupplies = 0
	}

	countOfBorrows, e := borrowStr.CountOfBorrows(ctx, market.Symbol)
	if e != nil {
		countOfBorrows = 0
	}

	marketView := views.Market{
		Market:          *market,
		ExchangeRate:    exchangeRate,
		UtilizationRate: utilizationRate,
		SupplyAPY:       supplyRate,
		BorrowAPY:       borrowRate,
		Suppliers:       countOfSupplies,
		Borrowers:       countOfBorrows,
	}

	return &marketView
}
