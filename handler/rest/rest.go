package rest

import (
	"compound/core"
	"compound/handler/render"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
)

// Handle handle rest api request
func Handle(system *core.System, dapp *core.Wallet, marketStore core.IMarketStore, supplyStore core.ISupplyStore, borrowStore core.IBorrowStore, transactionStore core.TransactionStore, oracleSignerStore core.OracleSignerStore, marketService core.IMarketService) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.NotFoundRequest(w, errors.New("not found"))
	})

	router.Get("/transactions", transactionsHandler(transactionStore))
	router.Get("/price-requests", priceRequestsHandler(system, marketStore, oracleSignerStore))
	router.Get("/markets/all", allMarketsHandler(marketStore, supplyStore, borrowStore, marketService))
	router.Post("/pay-requests", payRequestsHandler(system, dapp))

	return router
}
