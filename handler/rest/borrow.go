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

func borrowsHandler(ctx context.Context, marketStr core.IMarketStore, borrowStr core.IBorrowStore, priceSrv core.IPriceOracleService, blockSrv core.IBlockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			UserID string `json:"user"`
			Symbol string `json:"symbol"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		borrowViews := make([]*views.Borrow, 0)
		if params.Symbol != "" && params.UserID != "" {
			market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			borrow, e := borrowStr.Find(ctx, params.UserID, market.AssetID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			v := convert2BorrowView(ctx, market, borrow)

			borrowViews = append(borrowViews, v)
		} else if params.UserID != "" {
			borrows, e := borrowStr.FindByUser(ctx, params.UserID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, b := range borrows {
				market, e := marketStr.Find(ctx, b.AssetID)
				if e != nil {
					continue
				}

				v := convert2BorrowView(ctx, market, b)

				borrowViews = append(borrowViews, v)
			}
		} else if params.Symbol != "" {
			market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			borrows, e := borrowStr.FindByAssetID(ctx, market.AssetID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, b := range borrows {
				v := convert2BorrowView(ctx, market, b)

				borrowViews = append(borrowViews, v)
			}
		} else {
			//all
			borrows, e := borrowStr.All(ctx)
			if e != nil {
				render.BadRequest(w, e)
			}

			for _, b := range borrows {
				market, e := marketStr.Find(ctx, b.AssetID)
				if e != nil {
					continue
				}
				v := convert2BorrowView(ctx, market, b)

				borrowViews = append(borrowViews, v)
			}
		}

		render.JSON(w, borrowViews)
	}
}

func convert2BorrowView(ctx context.Context, market *core.Market, borrow *core.Borrow) *views.Borrow {
	borrowView := views.Borrow{
		Borrow: *borrow,
	}

	return &borrowView
}
