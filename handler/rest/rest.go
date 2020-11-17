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

	// MUST
	router.Get("/markets", allMarketsHandler(ctx, marketStore, supplyStore, borrowStore, marketService))
	// MUST
	router.Get("/markets/{symbol}", marketHandler(ctx, marketStore, supplyStore, borrowStore, marketService))
	// MUST
	router.Get("/accounts/{user}", accountHandler(ctx, blockService, accountService))

	// MUST
	// supplies?user=xxxxx&symbol=BTC
	router.Get("/supplies", suppliesHandler(ctx, marketStore, supplyStore, priceService, blockService))
	// create supply
	router.Post("/supplies", supplyHandler(ctx, marketStore, supplyService))
	// pledge
	router.Post("/supplies/pledge", pledgeHandler(ctx, marketStore, supplyService))
	router.Put("/supplies/unpledge", unpledgeHandler(ctx, marketStore, supplyService))
	router.Put("/supplies/redeem", redeemHandler(ctx, marketStore, supplyService))

	// MUST
	// borrows?user=xxxxx&symbol=BTC
	router.Get("/borrows", borrowsHandler(ctx, marketStore, borrowStore, priceService, blockService))
	// create borrow
	router.Post("/borrows", borrowHandler(ctx, marketStore, borrowService))
	router.Put("/borrows/repay", repayHandler(ctx, marketStore, borrowService))
	// borrows/max?user=xxxx&symbol=BTC
	router.Get("/borrows/max", nil)
	// borrows/repay/max?user=xxxxx&symbol=BTC
	router.Get("/borrows/repay/max", nil)

	// MUST
	// liquidities?user=xxxxx
	router.Get("/liquidities", liquiditiesHandler(ctx, accountService))
	router.Post("/seizetoken", seizeTokenHandler(ctx, marketStore, supplyStore, borrowStore, accountService))
	router.Get("/seizetoken/max", nil)

	return router
}
