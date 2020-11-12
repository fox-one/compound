package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"
)

func accountHandler(ctx context.Context, accountSrv core.IAccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			UserID string `json:"user"`
		}

		if err := param.Binding(r, &params); err != nil {
			render.BadRequest(w, err)
			return
		}

		liquidity, e := accountSrv.CalculateAccountLiquidity(ctx, params.UserID)
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
