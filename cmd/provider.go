package cmd

import (
	"compound/core"
	"compound/pkg/mtg"
	accountservice "compound/service/account"
	"compound/service/block"
	borrowservice "compound/service/borrow"
	marketservice "compound/service/market"
	oracle "compound/service/oracle"
	supplyservice "compound/service/supply"
	walletservice "compound/service/wallet"
	"compound/store/borrow"
	"compound/store/market"
	"compound/store/supply"
	"fmt"

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

func provideDapp() *core.Wallet {
	c, err := mixin.NewFromKeystore(&cfg.Dapp.Keystore)
	if err != nil {
		panic(err)
	}
	return &core.Wallet{
		Client: c,
		Pin:    cfg.Dapp.Pin,
	}
}

func provideSystem() *core.System {
	members := make([]*core.Member, 0, len(cfg.Group.Members))
	for _, m := range cfg.Group.Members {
		verifyKey, err := mtg.DecodePublicKey(m.VerifyKey)
		if err != nil {
			panic(fmt.Errorf("decode verify key for member %s failed", m.ClientID))
		}

		members = append(members, &core.Member{
			ClientID:  m.ClientID,
			VerifyKey: verifyKey,
		})
	}

	privateKey, err := mtg.DecodePrivateKey(cfg.Group.PrivateKey)
	if err != nil {
		panic(fmt.Errorf("base64 decode group private key failed: %w", err))
	}

	signKey, err := mtg.DecodePrivateKey(cfg.Group.SignKey)
	if err != nil {
		panic(fmt.Errorf("base64 decode group sign key failed: %w", err))
	}

	return &core.System{
		Admins:       cfg.Group.Admins,
		ClientID:     cfg.Dapp.ClientID,
		ClientSecret: cfg.Dapp.ClientSecret,
		Members:      members,
		Threshold:    cfg.Group.Threshold,
		VoteAsset:    cfg.Group.Vote.Asset,
		VoteAmount:   cfg.Group.Vote.Amount,
		PrivateKey:   privateKey,
		SignKey:      signKey,
		Location:     cfg.Location,
		Genesis:      cfg.Genesis,
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

// func provideTransferStore(db *db.DB) core.ITransferStore {
// 	return transfer.New(db)
// }

// func provideSnapshotStore(db *db.DB) core.ISnapshotStore {
// 	return snapshot.New(db)
// }

// ------------------service------------------------------------
func provideWalletService(client *mixin.Client, cfg walletservice.Config) core.WalletService {
	// return wallet.New(mainWallet)
	return nil
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

func provideSupplyService(db *db.DB, mainWallet *core.Wallet, blockWallet *core.Wallet, supplyStr core.ISupplyStore, marketStr core.IMarketStore, accountSrv core.IAccountService, priceSrv core.IPriceOracleService, blockSrv core.IBlockService, marketSrv core.IMarketService) core.ISupplyService {
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
		marketSrv,
	)
}

func provideBorrowService(mainWallet *core.Wallet, blockWallet *core.Wallet, marketStr core.IMarketStore, borrowStr core.IBorrowStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService, accountSrv core.IAccountService, marketSrv core.IMarketService) core.IBorrowService {
	return borrowservice.New(
		&cfg,
		mainWallet,
		blockWallet,
		marketStr,
		borrowStr,
		blockSrv,
		priceSrv,
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
	marketSrv core.IMarketService) core.IAccountService {

	return accountservice.New(mainWallet, marketStore, supplyStore, borrowStore, priceSrv, blockSrv, marketSrv)
}
