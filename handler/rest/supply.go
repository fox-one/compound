package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
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

		blockNum, e := blockSrv.GetBlock(ctx, time.Now())
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
			v, e := convert2SupplyView(ctx, market, supply, blockNum, priceSrv)
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

				v, e := convert2SupplyView(ctx, market, s, blockNum, priceSrv)
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
				v, e := convert2SupplyView(ctx, market, s, blockNum, priceSrv)
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
				v, e := convert2SupplyView(ctx, market, s, blockNum, priceSrv)
				if e != nil {
					continue
				}

				supplyViews = append(supplyViews, v)
			}
		}

		render.JSON(w, supplyViews)
	}
}

func supplyHandler(ctx context.Context, marketStr core.IMarketStore, supplySrv core.ISupplyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Symbol string          `json:"symbol"`
			Amount decimal.Decimal `json:"amount"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		payURL, e := supplySrv.Supply(ctx, params.Amount, market)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		rsp := views.PayURL{
			PayURL: payURL,
		}

		render.JSON(w, &rsp)
	}
}

func pledgeHandler(ctx context.Context, marketStr core.IMarketStore, supplySrv core.ISupplyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Symbol string          `json:"symbol"`
			Amount decimal.Decimal `json:"amount"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		payURL, e := supplySrv.Pledge(ctx, params.Amount, market)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		rsp := views.PayURL{
			PayURL: payURL,
		}

		render.JSON(w, &rsp)
	}
}

func unpledgeHandler(ctx context.Context, marketStr core.IMarketStore, supplySrv core.ISupplyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			UserID string          `json:"user"`
			Symbol string          `json:"symbol"`
			Amount decimal.Decimal `json:"amount"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		e = supplySrv.Unpledge(ctx, params.Amount, params.UserID, market)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		render.JSON(w, &views.DefaultSuccess)
	}
}

func redeemHandler(ctx context.Context, marketStr core.IMarketStore, supplySrv core.ISupplyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Symbol string          `json:"symbol"`
			Amount decimal.Decimal `json:"amount"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		url, e := supplySrv.Redeem(ctx, params.Amount, market)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		rsp := views.PayURL{
			PayURL: url,
		}

		render.JSON(w, &rsp)
	}
}

func convert2SupplyView(ctx context.Context, market *core.Market, supply *core.Supply, curBlock int64, priceSrv core.IPriceOracleService) (*views.Supply, error) {
	price, e := priceSrv.GetCurrentUnderlyingPrice(ctx, market)
	if e != nil {
		return nil, e
	}

	supplyView := views.Supply{
		Supply: *supply,
		Price:  price,
	}

	return &supplyView, nil
}
