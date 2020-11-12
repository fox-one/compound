package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"
	"strings"
)

func suppliesHandler(ctx context.Context, marketStr core.IMarketStore, supplyStr core.ISupplyStore, priceSrv core.IPriceOracleService, blockSrv core.IBlockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			UserID string `json:"user"`
			Symbol string `json:"symbol"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		curBlock, e := blockSrv.CurrentBlock(ctx)
		if e != nil {
			render.BadRequest(w, e)
			return
		}
		supplyViews := make([]*views.Supply, 0)
		if params.Symbol != "" && params.UserID != "" {
			market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			supply, e := supplyStr.Find(ctx, params.UserID, market.CTokenAssetID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			v, e := convert2SupplyView(ctx, market, supply, curBlock, priceSrv)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			supplyViews = append(supplyViews, v)
		} else if params.UserID != "" {
			supplies, e := supplyStr.FindByUser(ctx, params.UserID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, s := range supplies {
				market, e := marketStr.FindByCToken(ctx, s.CTokenAssetID)
				if e != nil {
					continue
				}

				v, e := convert2SupplyView(ctx, market, s, curBlock, priceSrv)
				if e != nil {
					continue
				}

				supplyViews = append(supplyViews, v)
			}
		} else if params.Symbol != "" {
			market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
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
				market, e := marketStr.FindByCToken(ctx, s.CTokenAssetID)
				if e != nil {
					continue
				}
				v, e := convert2SupplyView(ctx, market, s, curBlock, priceSrv)
				if e != nil {
					continue
				}

				supplyViews = append(supplyViews, v)
			}
		} else {
			//all
			supplies, e := supplyStr.All(ctx)
			if e != nil {
				render.BadRequest(w, e)
			}

			for _, s := range supplies {
				market, e := marketStr.FindByCToken(ctx, s.CTokenAssetID)
				if e != nil {
					continue
				}
				v, e := convert2SupplyView(ctx, market, s, curBlock, priceSrv)
				if e != nil {
					continue
				}

				supplyViews = append(supplyViews, v)
			}
		}

		render.JSON(w, supplyViews)
	}
}

func convert2SupplyView(ctx context.Context, market *core.Market, supply *core.Supply, curBlock int64, priceSrv core.IPriceOracleService) (*views.Supply, error) {
	price, e := priceSrv.GetUnderlyingPrice(ctx, market.Symbol, curBlock)
	if e != nil {
		return nil, e
	}

	supplyView := views.Supply{
		Supply: *supply,
		Price:  price,
	}

	return &supplyView, nil
}
