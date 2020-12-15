package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"net/http"
)

func suppliesHandler(marketStr core.IMarketStore, supplyStr core.ISupplyStore, priceSrv core.IPriceOracleService, blockSrv core.IBlockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var params struct {
			UserID string `json:"user"`
			Asset  string `json:"asset"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		supplyViews := make([]*views.Supply, 0)
		if params.Asset != "" && params.UserID != "" {
			market, _, e := marketStr.Find(ctx, params.Asset)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			supply, _, e := supplyStr.Find(ctx, params.UserID, market.CTokenAssetID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			v := convert2SupplyView(market, supply)

			supplyViews = append(supplyViews, v)
		} else if params.UserID != "" {
			supplies, e := supplyStr.FindByUser(ctx, params.UserID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, s := range supplies {
				market, _, e := marketStr.FindByCToken(ctx, s.CTokenAssetID)
				if e != nil {
					continue
				}

				v := convert2SupplyView(market, s)

				supplyViews = append(supplyViews, v)
			}
		} else if params.Asset != "" {
			market, _, e := marketStr.Find(ctx, params.Asset)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			supplies, e := supplyStr.FindByCTokenAssetID(ctx, market.CTokenAssetID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, s := range supplies {
				v := convert2SupplyView(market, s)

				supplyViews = append(supplyViews, v)
			}
		} else {
			//all
			supplies, e := supplyStr.All(ctx)
			if e != nil {
				render.BadRequest(w, e)
			}

			for _, s := range supplies {
				market, _, e := marketStr.FindByCToken(ctx, s.CTokenAssetID)
				if e != nil {
					continue
				}
				v := convert2SupplyView(market, s)

				supplyViews = append(supplyViews, v)
			}
		}

		render.JSON(w, supplyViews)
	}
}

func convert2SupplyView(market *core.Market, supply *core.Supply) *views.Supply {
	supplyView := views.Supply{
		Supply: *supply,
	}

	return &supplyView
}
