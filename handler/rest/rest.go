package rest

import (
	"compound/core"
	"compound/handler/render"
	"context"
	"net/http"

	"github.com/fox-one/pkg/store/db"
	"github.com/go-chi/chi"
	"github.com/twitchtv/twirp"
)

// Handle handle rest request
func Handle(ctx context.Context,
	config *core.Config,
	db *db.DB,
	mainWallet *core.Wallet,
	blockWallet *core.Wallet,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	blockService core.IBlockService,
	priceService core.IPriceOracleService,
	accountService core.IAccountService,
	marketService core.IMarketService,
	supplyService core.ISupplyService,
	borrowService core.IBorrowService) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, twirp.NotFoundError("not found"))
	})

	router.Get("/markets", allMarketsHandler(ctx, marketStore, supplyStore, borrowStore, marketService))
	router.Get("/markets/{asset}", marketHandler(ctx, marketStore, supplyStore, borrowStore, marketService))
	router.Get("/liquidities/{user}", liquidityHandler(ctx, blockService, accountService))

	// supplies?user=xxxxx&asset=xxxxx
	router.Get("/supplies", suppliesHandler(ctx, marketStore, supplyStore, priceService, blockService))
	// borrows?user=xxxxx&asset=xxxx
	router.Get("/borrows", borrowsHandler(ctx, marketStore, borrowStore, priceService, blockService))

	return router
}
