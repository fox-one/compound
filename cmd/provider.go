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
	"compound/store/borrow"
	"compound/store/market"
	"compound/store/supply"
	"compound/store/transfer"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/property"
	"github.com/fox-one/pkg/store/db"
	propertystore "github.com/fox-one/pkg/store/property"
)

func provideDatabase() *db.DB {
	return db.MustOpen(cfg.DB)
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
	c, err := mixin.NewFromKeystore(&cfg.GasWallet.Keystore)
	if err != nil {
		panic(err)
	}

	return &core.Wallet{
		Client: c,
		Pin:    provideConfig().GasWallet.Pin,
	}
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

func provideTransferStore(db *db.DB) core.ITransferStore {
	return transfer.New(db)
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

func provideMarketService(mainWallet *core.Wallet, marketStr core.IMarketStore, borrowStr core.IBorrowStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService) core.IMarketService {
	return marketservice.New(
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
	priceSrv core.IPriceOracleService,
	blockSrv core.IBlockService,
	walletService core.IWalletService,
	marketSrv core.IMarketService) core.IAccountService {

	return accountservice.New(mainWallet, marketStore, supplyStore, borrowStore, priceSrv, blockSrv, walletService, marketSrv)
}
