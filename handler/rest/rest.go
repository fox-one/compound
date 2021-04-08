package rest

import (
	"compound/core"
	"compound/handler/render"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/twitchtv/twirp"
)

// Handle handle rest api request
func Handle(config *core.Config, marketStore core.IMarketStore, transactionStore core.TransactionStore) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, twirp.NotFoundError("not found"))
	})

	// marketStore :=

	router.Get("/transactions", transactionsHandler(transactionStore))
	router.Get("/price-requests", priceRequestsHandler(config, marketStore))

	return router
}
