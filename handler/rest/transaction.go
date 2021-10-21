package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"net/http"
	"time"
)

// response user transactions
func transactionsHandler(transactionStr core.TransactionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var params struct {
			Offset string `json:"offset"`
			Limit  int    `json:"limit"`
		}

		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		limit := params.Limit
		if limit <= 0 {
			limit = 500
		}

		offsetTime, err := time.Parse(time.RFC3339Nano, params.Offset)
		if err != nil {
			offsetTime = time.Time{}
		}

		transactions, e := transactionStr.List(ctx, offsetTime, limit)
		if e != nil {
			render.BadRequest(w, e)
			return
		}

		render.JSON(w, transactions)
	}
}
