package cmd

import (
	"compound/core"
	"compound/service/block"
	marketservice "compound/service/market"
	oracle "compound/service/oracle"
	"compound/service/wallet"
	"compound/store/market"
	"compound/store/user"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/property"
	"github.com/fox-one/pkg/store/db"
	propertystore "github.com/fox-one/pkg/store/property"

	"github.com/go-redis/redis"
)

func provideDatabase() *db.DB {
	return db.MustOpen(cfg.DB)
}

func provideRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
		DB:   cfg.Redis.DB,
	})
}

func provideConfig() *core.Config {
	return &cfg
}

func provideDApp() *mixin.Client {
	c, err := mixin.NewFromKeystore(&cfg.Mixin.Keystore)
	if err != nil {
		panic(err)
	}

	return c
}

func provideMainWallet() *mixin.Client {
	return provideDApp()
} 

func provideBlockWallet() *mixin.Client {
	c, err := mixin.NewFromKeystore(&cfg.BlockWallet.Keystore)
	if err != nil {
		panic(err)
	}

	return c
}

func provideReserveWallet() *mixin.Client {
	c, err := mixin.NewFromKeystore(&cfg.BlockWallet.Keystore)
	if err != nil {
		panic(err)
	}

	return c
}

// ---------------store-----------------------------------------

func provideUserStore(db *db.DB) core.IUserStore {
	return user.New(db)
}

func providePropertyStore(db *db.DB) property.Store {
	return propertystore.New(db)
}

func provideMarketStore() core.IMarketStore {
	return market.New(provideDatabase())
}

// ------------------service------------------------------------
func provideWalletService() core.IWalletService {
	return wallet.New(provideDApp(), cfg.Mixin.Pin)
}

func provideBlockService() core.IBlockService {
	return block.New(provideConfig())
}

func providePriceService() core.IPriceOracleService {
	return oracle.New(provideConfig(),
		provideRedis(),
		provideBlockService())
}

func provideMarketService() core.IMarketService {
	return marketservice.New(provideRedis(),
		provideDApp(),
		provideMarketStore(),
		provideBlockService(),
		providePriceService())
}
