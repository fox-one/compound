package cmd

import (
	"compound/core"
	"compound/store/user"

	"github.com/fox-one/pkg/store/db"

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

func provideUserStore(db *db.DB) core.IUserStore {
	return user.New(db)
}
