package rest

import (
	"compound/core"
	"compound/handler/render"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/twitchtv/twirp"
)

// Handle handle rest api request
func Handle(system *core.System, marketStore core.IMarketStore, supplyStore core.ISupplyStore, borrowStore core.IBorrowStore, transactionStore core.TransactionStore, marketService core.IMarketService) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, twirp.NotFoundError("not found"))
	})

	router.Get("/transactions", transactionsHandler(transactionStore))
	router.Get("/price-requests", priceRequestsHandler(system, marketStore))
	router.Get("/markets/all", allMarketsHandler(marketStore, supplyStore, borrowStore, marketService))

	return router
}
