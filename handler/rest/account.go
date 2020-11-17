package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"
	"time"
)

func accountHandler(ctx context.Context, blockSrv core.IBlockService, accountSrv core.IAccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			UserID string `json:"user"`
		}

		if err := param.Binding(r, &params); err != nil {
			render.BadRequest(w, err)
			return
		}

		blockNum, e := blockSrv.GetBlock(ctx, time.Now())
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		liquidity, e := accountSrv.CalculateAccountLiquidity(ctx, params.UserID, blockNum)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		accountView := views.Account{
			Liquidity: liquidity,
		}

		render.JSON(w, &accountView)
	}
}
