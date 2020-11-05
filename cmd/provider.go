package cmd

import (
	"compound/core"
	"compound/service/block"
	marketservice "compound/service/market"
	oracle "compound/service/oracle"
	"compound/service/wallet"
	"compound/store/borrow"
	"compound/store/market"
	"compound/store/supply"
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

func provideMainWallet() *core.Wallet {
	c, err := mixin.NewFromKeystore(&cfg.MainWallet.Keystore)
	if err != nil {
		panic(err)
	}
	return &core.Wallet{
		Client: c,
		Pin:    provideConfig().MainWallet.Pin,
	}
}

func provideBlockWallet() *core.Wallet {
	c, err := mixin.NewFromKeystore(&cfg.BlockWallet.Keystore)
	if err != nil {
		panic(err)
	}

	return &core.Wallet{
		Client: c,
		Pin:    provideConfig().BlockWallet.Pin,
	}
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

func provideSupplyStore() core.ISupplyStore {
	return supply.New(provideDatabase())
}

func provideBorrowStore() core.IBorrowStore {
	return borrow.New(provideDatabase())
}

// ------------------service------------------------------------
func provideWalletService() core.IWalletService {
	return wallet.New(provideMainWallet())
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
		provideMainWallet(),
		provideMarketStore(),
		provideBorrowStore(),
		provideBlockService(),
		providePriceService())
}

func provideSupplyService() core.ISupplyService {
	// return supplyService.New(provideConfig(), provideMainWallet())
	return nil
}

func provideBorrowService() core.IBorrowService {
	// return borrowService.New()
	return nil
}

func provideAccountService() core.IAccountService {
	return nil
}
