package cmd

import (
	"compound/core"
	accountservice "compound/service/account"
	"compound/service/block"
	borrowservice "compound/service/borrow"
	marketservice "compound/service/market"
	oracle "compound/service/oracle"
	supplyservice "compound/service/supply"
	"compound/service/wallet"
	"compound/store/account"
	"compound/store/borrow"
	"compound/store/market"
	"compound/store/supply"

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

//TODO 不单独提供保留金钱包，以记账方式保存数据库记录
func provideReserveWallet() *mixin.Client {
	c, err := mixin.NewFromKeystore(&cfg.BlockWallet.Keystore)
	if err != nil {
		panic(err)
	}

	return c
}

// ---------------store-----------------------------------------
func providePropertyStore(db *db.DB) property.Store {
	return propertystore.New(db)
}

func provideMarketStore(db *db.DB) core.IMarketStore {
	return market.New(db)
}

func provideSupplyStore(db *db.DB) core.ISupplyStore {
	return supply.New(db)
}

func provideBorrowStore(db *db.DB) core.IBorrowStore {
	return borrow.New(db)
}

func provideAccountStore(redis *redis.Client) core.IAccountStore {
	return account.New(redis)
}

// ------------------service------------------------------------
func provideWalletService(mainWallet *core.Wallet) core.IWalletService {
	return wallet.New(mainWallet)
}

func provideBlockService() core.IBlockService {
	return block.New(&cfg)
}

func providePriceService(blockSrv core.IBlockService) core.IPriceOracleService {
	return oracle.New(&cfg, blockSrv)
}

func provideMarketService(redis *redis.Client, mainWallet *core.Wallet, marketStr core.IMarketStore, borrowStr core.IBorrowStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService) core.IMarketService {
	return marketservice.New(redis,
		mainWallet,
		marketStr,
		borrowStr,
		blockSrv,
		priceSrv)
}

func provideSupplyService(db *db.DB, mainWallet *core.Wallet, blockWallet *core.Wallet, supplyStr core.ISupplyStore, marketStr core.IMarketStore, accountSrv core.IAccountService, priceSrv core.IPriceOracleService, blockSrv core.IBlockService, walletSrv core.IWalletService, marketSrv core.IMarketService) core.ISupplyService {
	return supplyservice.New(
		&cfg,
		db,
		mainWallet,
		blockWallet,
		supplyStr,
		marketStr,
		accountSrv,
		priceSrv,
		blockSrv,
		walletSrv,
		marketSrv,
	)
}

func provideBorrowService(mainWallet *core.Wallet, blockWallet *core.Wallet, marketStr core.IMarketStore, borrowStr core.IBorrowStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService, walletSrv core.IWalletService, accountSrv core.IAccountService, marketSrv core.IMarketService) core.IBorrowService {
	return borrowservice.New(
		&cfg,
		mainWallet,
		blockWallet,
		marketStr,
		borrowStr,
		blockSrv,
		priceSrv,
		walletSrv,
		accountSrv,
		marketSrv,
	)
}

func provideAccountService(mainWallet *core.Wallet,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	accountStore core.IAccountStore,
	priceSrv core.IPriceOracleService,
	blockSrv core.IBlockService,
	walletService core.IWalletService,
	marketSrv core.IMarketService) core.IAccountService {

	return accountservice.New(mainWallet, marketStore, supplyStore, borrowStore, accountStore, priceSrv, blockSrv, walletService, marketSrv)
}
