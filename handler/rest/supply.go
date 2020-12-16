package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"net/http"
)

func suppliesHandler(userStr core.UserStore, marketStr core.IMarketStore, supplyStr core.ISupplyStore, priceSrv core.IPriceOracleService, blockSrv core.IBlockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var params struct {
			Address string `json:"address"`
			Asset   string `json:"asset"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		supplyViews := make([]*views.Supply, 0)
		if params.Asset != "" && params.Address != "" {
			market, _, e := marketStr.Find(ctx, params.Asset)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			user, e := userStr.FindByAddress(ctx, params.Address)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			supply, _, e := supplyStr.Find(ctx, user.UserID, market.CTokenAssetID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			v := convert2SupplyView(user.Address, supply)

			supplyViews = append(supplyViews, v)
		} else if params.Address != "" {
			user, e := userStr.FindByAddress(ctx, params.Address)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			supplies, e := supplyStr.FindByUser(ctx, user.UserID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, s := range supplies {
				v := convert2SupplyView(user.Address, s)

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
				v := convert2SupplyView(core.BuildUserAddress(s.UserID), s)

				supplyViews = append(supplyViews, v)
			}
		} else {
			//all
			supplies, e := supplyStr.All(ctx)
			if e != nil {
				render.BadRequest(w, e)
			}

			for _, s := range supplies {
				v := convert2SupplyView(core.BuildUserAddress(s.UserID), s)

				supplyViews = append(supplyViews, v)
			}
		}

		render.JSON(w, supplyViews)
	}
}

func convert2SupplyView(address string, supply *core.Supply) *views.Supply {
	supplyView := views.Supply{
		Supply:  *supply,
		Address: address,
	}

	return &supplyView
}
