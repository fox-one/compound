package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"context"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
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

		curBlock, e := blockSrv.CurrentBlock(ctx)
		if e != nil {
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
			borrows, e := borrowStr.Find(ctx, params.UserID, params.Symbol)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, b := range borrows {
				v, e := convert2BorrowView(ctx, market, b, curBlock, priceSrv)
				if e != nil {
					continue
				}

				borrowViews = append(borrowViews, v)
			}
		} else if params.UserID != "" {
			borrows, e := borrowStr.FindByUser(ctx, params.UserID)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, b := range borrows {
				market, e := marketStr.FindBySymbol(ctx, b.Symbol)
				if e != nil {
					continue
				}

				v, e := convert2BorrowView(ctx, market, b, curBlock, priceSrv)
				if e != nil {
					continue
				}

				borrowViews = append(borrowViews, v)
			}
		} else if params.Symbol != "" {
			market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
			if e != nil {
				render.BadRequest(w, e)
				return
			}
			borrows, e := borrowStr.FindBySymbol(ctx, market.Symbol)
			if e != nil {
				render.BadRequest(w, e)
				return
			}

			for _, b := range borrows {
				v, e := convert2BorrowView(ctx, market, b, curBlock, priceSrv)
				if e != nil {
					continue
				}

				borrowViews = append(borrowViews, v)
			}
		} else {
			//all
			borrows, e := borrowStr.All(ctx)
			if e != nil {
				render.BadRequest(w, e)
			}

			for _, b := range borrows {
				market, e := marketStr.FindBySymbol(ctx, b.Symbol)
				if e != nil {
					continue
				}
				v, e := convert2BorrowView(ctx, market, b, curBlock, priceSrv)
				if e != nil {
					continue
				}

				borrowViews = append(borrowViews, v)
			}
		}

		render.JSON(w, borrowViews)
	}
}

func borrowHandler(ctx context.Context, marketStr core.IMarketStore, borrowSrv core.IBorrowService) http.HandlerFunc {
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

		e = borrowSrv.Borrow(ctx, params.Amount, params.UserID, market)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		render.JSON(w, views.DefaultSuccess)
	}
}

func repayHandler(ctx context.Context, marketStr core.IMarketStore, borrowSrv core.IBorrowService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var params struct {
		// 	Symbol string          `json:"symbol"`
		// 	Amount decimal.Decimal `json:"amount"`
		// }

		// if e := param.Binding(r, &params); e != nil {
		// 	render.BadRequest(w, e)
		// 	return
		// }

		// market, e := marketStr.FindBySymbol(ctx, strings.ToUpper(params.Symbol))
		// if e != nil {
		// 	render.BadRequest(w, e)
		// 	return
		// }

		// url, e := borrowSrv.Repay(ctx, params.Amount, market)
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

func convert2BorrowView(ctx context.Context, market *core.Market, borrow *core.Borrow, curBlock int64, priceSrv core.IPriceOracleService) (*views.Borrow, error) {
	price, e := priceSrv.GetUnderlyingPrice(ctx, market.Symbol, curBlock)
	if e != nil {
		return nil, e
	}

	borrowView := views.Borrow{
		Borrow: *borrow,
		Price:  price,
	}

	return &borrowView, nil
}
