package cmd

import (
	"compound/core"
	"compound/service/wallet"
	"compound/store/user"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
	"gopkg.in/redis.v5"
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

func provideUserStore(db *db.DB) core.IUserStore {
	return user.New(db)
}

func provideMixinClient() *mixin.Client {
	c, err := mixin.NewFromKeystore(&cfg.Mixin.Keystore)
	if err != nil {
		panic(err)
	}

	return c
}

func provideConfig() *core.Config {
	return &cfg
}

func provideWalletService() core.IWalletService {
	return wallet.New(provideMixinClient(), cfg.Mixin.Pin)
}
