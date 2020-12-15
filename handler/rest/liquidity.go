package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"net/http"
	"time"
)

func liquidityHandler(blockSrv core.IBlockService, accountSrv core.IAccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var params struct {
			UserID string `json:"user"`
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
		liqudity, e := accountSrv.CalculateAccountLiquidity(ctx, params.UserID, blockNum)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		accountView := views.Account{
			Liquidity: liqudity,
		}

		render.JSON(w, accountView)
	}
}
