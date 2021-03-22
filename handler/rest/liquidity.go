package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"net/http"
	"time"
)

// response liquidity by address
func liquidityHandler(userStr core.UserStore, blockSrv core.IBlockService, accountSrv core.IAccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var params struct {
			Address string `json:"address"`
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

		user, e := userStr.FindByAddress(ctx, params.Address)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		liqudity, e := accountSrv.CalculateAccountLiquidity(ctx, user.UserID, blockNum)
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
