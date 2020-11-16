package rest

import (
	"compound/core"
	"compound/handler/render"
	"context"
	"net/http"
)

func liquiditiesHandler(ctx context.Context, accountSrv core.IAccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accounts, e := accountSrv.SeizeAllowedAccounts(ctx)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		render.JSON(w, accounts)
	}
}

func seizeTokenHandler(ctx context.Context, marketStr core.IMarketStore, supplyStr core.ISupplyStore, borrowStr core.IBorrowStore, accountSrv core.IAccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var params struct {
		// 	UserID      string          `json:"user"`
		// 	SeizeSymbol string          `json:"seize_symbol"`
		// 	RepaySymbol string          `json:"repay_symbol"`
		// 	RepayAmount decimal.Decimal `json:"repay_amount"`
		// }

		// if e := param.Binding(r, &params); e != nil {
		// 	render.BadRequest(w, e)
		// 	return
		// }

		// market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.SeizeSymbol))
		// if e != nil {
		// 	render.BadRequest(w, e)
		// 	return
		// }

		// supply, e := supplyStr.Find(ctx, params.UserID, market.CTokenAssetID)
		// if e != nil {
		// 	render.BadRequest(w, e)
		// 	return
		// }

		// borrow, e := borrowStr.Find(ctx, params.UserID, params.RepaySymbol)
		// if e != nil {
		// 	render.BadRequest(w, e)
		// 	return
		// }

		// url, e := accountSrv.SeizeToken(ctx, supply, borrow, params.RepayAmount)
		// if e != nil {
		// 	render.BadRequest(w, e)
		// 	return
		// }

		// payURL := views.PayURL{
		// 	PayURL: url,
		// }

		render.JSON(w, nil)
	}
}
