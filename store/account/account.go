package account

import (
	"compound/core"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/shopspring/decimal"
)

type accountStore struct {
	Redis *redis.Client
}

// New new account store
func New(redis *redis.Client) core.IAccountStore {
	return &accountStore{
		Redis: redis,
	}
}

func (s *accountStore) SaveLiquidity(ctx context.Context, userID string, block int64, liquidity decimal.Decimal) error {
	k := s.liquidityCacheKey(userID, block)

	if s.Redis.Exists(k).Val() == 0 {
		s.Redis.Set(k, []byte(liquidity.String()), time.Hour)
	}
	return nil
}
func (s *accountStore) FindLiquidity(ctx context.Context, userID string, curBlock int64) (decimal.Decimal, error) {
	k := s.liquidityCacheKey(userID, curBlock)
	bs, e := s.Redis.Get(k).Bytes()
	if e != nil {
		return decimal.Zero, e
	}
	liquidity, e := decimal.NewFromString(string(bs))
	if e != nil {
		return decimal.Zero, e
	}

	return liquidity, nil
}

func (s *accountStore) liquidityCacheKey(userID string, block int64) string {
	return fmt.Sprintf("foxone:compound:liqudity:%s:%d", userID, block)
}
