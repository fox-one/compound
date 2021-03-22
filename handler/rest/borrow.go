package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"net/http"
)

// response borrows by address and asset
func borrowsHandler(userStr core.UserStore, marketStr core.IMarketStore, borrowStr core.IBorrowStore, priceSrv core.IPriceOracleService, blockSrv core.IBlockService) http.HandlerFunc {
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

		borrowViews := make([]*views.Borrow, 0)
		if params.Asset != "" && params.Address != "" {
			market, _, e := marketStr.FindBySymbol(ctx, params.Asset)
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			user, e := userStr.FindByAddress(ctx, params.Address)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			borrow, _, e := borrowStr.Find(ctx, user.UserID, market.AssetID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			v := convert2BorrowView(user.Address, borrow)

			borrowViews = append(borrowViews, v)
		} else if params.Address != "" {
			user, e := userStr.FindByAddress(ctx, params.Address)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			borrows, e := borrowStr.FindByUser(ctx, user.UserID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, b := range borrows {
				v := convert2BorrowView(user.Address, b)

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
				v := convert2BorrowView(core.BuildUserAddress(b.UserID), b)

				borrowViews = append(borrowViews, v)
			}
		} else {
			//all
			borrows, e := borrowStr.All(ctx)
			if e != nil {
				render.BadRequest(w, e)
			}

			for _, b := range borrows {
				v := convert2BorrowView(core.BuildUserAddress(b.UserID), b)

				borrowViews = append(borrowViews, v)
			}
		}

		render.JSON(w, borrowViews)
	}
}

func convert2BorrowView(address string, borrow *core.Borrow) *views.Borrow {
	borrowView := views.Borrow{
		Borrow:  *borrow,
		Address: address,
	}

	return &borrowView
}
