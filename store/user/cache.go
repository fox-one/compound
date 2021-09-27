package user

import (
	"compound/core"
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"golang.org/x/sync/singleflight"
)

func Cache(store core.UserStore, exp time.Duration) core.UserStore {
	return &cacheUserStore{
		UserStore: store,
		cache:     gcache.New(2048).LRU().Build(),
		sf:        &singleflight.Group{},
	}
}

type cacheUserStore struct {
	core.UserStore
	cache gcache.Cache
	sf    *singleflight.Group
}

func (s *cacheUserStore) List(ctx context.Context, from uint64, limit int) ([]*core.User, error) {
	return s.UserStore.List(ctx, from, limit)
}

func (s *cacheUserStore) Create(ctx context.Context, user *core.User) error {
	if err := s.UserStore.Create(ctx, user); err != nil {
		return err
	}
	s.cacheUser(user)
	return nil
}

func (s *cacheUserStore) MigrateToV1(ctx context.Context, users []*core.User) error {
	if err := s.UserStore.MigrateToV1(ctx, users); err != nil {
		return err
	}
	for _, user := range users {
		s.cacheUser(user)
	}
	return nil
}

func (s *cacheUserStore) Find(ctx context.Context, userID string) (*core.User, error) {
	if v, err := s.cache.Get(s.userKey(userID)); err == nil {
		if user, ok := v.(*core.User); ok {
			return user, nil
		}
	}
	user, err := s.UserStore.Find(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.ID > 0 {
		s.cacheUser(user)
	}
	return user, nil
}

func (s *cacheUserStore) FindByAddress(ctx context.Context, address string) (*core.User, error) {
	if v, err := s.cache.Get(s.addressKey(address)); err == nil {
		if user, ok := v.(*core.User); ok {
			return user, nil
		}
	}
	if v, err := s.cache.Get(s.v0AddressKey(address)); err == nil {
		if user, ok := v.(*core.User); ok {
			return user, nil
		}
	}
	user, err := s.UserStore.FindByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	if user.ID > 0 {
		s.cacheUser(user)
	}
	return user, nil
}

func (s *cacheUserStore) cacheUser(user *core.User) {
	s.cache.Set(s.userKey(user.UserID), user)
	s.cache.Set(s.addressKey(user.Address), user)
	s.cache.Set(s.v0AddressKey(user.AddressV0), user)
}

func (s *cacheUserStore) userKey(userID string) string {
	return fmt.Sprintf("user:id:%s", userID)
}

func (s *cacheUserStore) addressKey(address string) string {
	return fmt.Sprintf("user:address:%s", address)
}

func (s *cacheUserStore) v0AddressKey(address string) string {
	return fmt.Sprintf("user:v0:address:%s", address)
}
