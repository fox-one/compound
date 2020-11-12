package rest

import (
	"compound/core"
	"compound/handler/render"
	"context"
	"net/http"

	"github.com/fox-one/pkg/store/db"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"github.com/twitchtv/twirp"
)

// Handle handle rest request
func Handle(ctx context.Context,
	config *core.Config,
	db *db.DB,
	redis *redis.Client,
	mainWallet *core.Wallet,
	blockWallet *core.Wallet,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	accountStore core.IAccountStore,
	walletService core.IWalletService,
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
	router.Get("/markets/{symbol}", marketHandler(ctx, marketStore, supplyStore, borrowStore, marketService))

	router.Get("/accounts/{user}", accountHandler(ctx, accountService))

	// supplies?user=xxxxx&symbol=BTC
	router.Get("/supplies", suppliesHandler(ctx, marketStore, supplyStore, priceService, blockService))
	// create supply
	router.Post("/supplies", nil)
	// pledge
	router.Post("/supplies/pledge", nil)
	// supplies/pledge/max?user=xxx&symbol=BTC
	router.Get("/supplies/pledge/max", nil)
	router.Put("/supplies/unpledge", nil)
	router.Put("/supplies/redeem", nil)
	// supplies/redeem/max?user=xxxx&symbol=BTC
	router.Get("/supplies/redeem/max", nil)

	// borrows?user=xxxxx&symbol=BTC
	router.Get("/borrows", nil)
	// create borrow
	router.Post("/borrows", nil)
	// borrows/max?user=xxxx&symbol=BTC
	router.Get("/borrows/max", nil)
	router.Put("/borrows/repay", nil)
	// borrows/repay/max?user=xxxxx&symbol=BTC
	router.Get("/borrows/repay/max", nil)

	// liquidities?user=xxxxx
	router.Get("/liquidities", nil)
	router.Post("/seizetoken", nil)
	router.Get("/seizetoken/max", nil)

	return router
}
