package rest

import (
	"compound/core"
	"compound/handler/render"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/twitchtv/twirp"
)

// Handle handle rest api request
func Handle(
	userStore core.UserStore,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	transactionStore core.TransactionStore,
	blockService core.IBlockService,
	priceService core.IPriceOracleService,
	accountService core.IAccountService,
	marketService core.IMarketService) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, twirp.NotFoundError("not found"))
	})

	router.Get("/markets", allMarketsHandler(marketStore, supplyStore, borrowStore, marketService))
	router.Get("/markets/{asset}", marketHandler(marketStore, supplyStore, borrowStore, marketService))
	router.Get("/liquidities/{address}", liquidityHandler(userStore, blockService, accountService))

	// supplies?address=xxxxx&asset=xxxxx
	router.Get("/supplies", suppliesHandler(userStore, marketStore, supplyStore, priceService, blockService))
	// borrows?address=xxxxx&asset=xxxx
	router.Get("/borrows", borrowsHandler(userStore, marketStore, borrowStore, priceService, blockService))
	router.Get("/transactions", transactionsHandler(transactionStore))

	return router
}
