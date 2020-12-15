package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"
)

func borrowsHandler(ctx context.Context, marketStr core.IMarketStore, borrowStr core.IBorrowStore, priceSrv core.IPriceOracleService, blockSrv core.IBlockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			UserID string `json:"user"`
			Asset  string `json:"asset"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		borrowViews := make([]*views.Borrow, 0)
		if params.Asset != "" && params.UserID != "" {
			market, _, e := marketStr.FindBySymbol(ctx, params.Asset)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			borrow, _, e := borrowStr.Find(ctx, params.UserID, market.AssetID)
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
				market, _, e := marketStr.Find(ctx, b.AssetID)
				if e != nil {
					continue
				}

				v := convert2BorrowView(ctx, market, b)

				borrowViews = append(borrowViews, v)
			}
		} else if params.Asset != "" {
			market, _, e := marketStr.Find(ctx, params.Asset)
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
				market, _, e := marketStr.Find(ctx, b.AssetID)
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
