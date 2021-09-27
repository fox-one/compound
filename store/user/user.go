package user

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store"
	"github.com/fox-one/pkg/store/db"
)

type userStore struct {
	db *db.DB
}

// New new user store
func New(db *db.DB) core.UserStore {
	return &userStore{
		db: db,
	}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.User{})

		if err := tx.AutoMigrate(core.User{}).Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_users_user_id", "user_id").Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_users_address", "address").Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_users_address_v0", "address_v0").Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *userStore) List(ctx context.Context, from uint64, limit int) ([]*core.User, error) {
	var users []*core.User
	if err := s.db.View().Where("id > ?", from).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userStore) Create(ctx context.Context, user *core.User) error {
	return s.db.Update().Where("user_id = ?", user.UserID).FirstOrCreate(user).Error
}

func (s *userStore) MigrateToV1(ctx context.Context, users []*core.User) error {
	return s.db.Tx(func(tx *db.DB) error {
		for _, user := range users {
			tx := tx.Update().Model(user).Where("address_v0 IS NULL").
				Updates(map[string]interface{}{
					"address":    user.Address,
					"address_v0": user.AddressV0,
				})
			if tx.Error != nil {
				return tx.Error
			}

			if tx.RowsAffected == 0 {
				return db.ErrOptimisticLock
			}
		}
		return nil
	})
}

func (s *userStore) Find(ctx context.Context, mixinUserID string) (*core.User, error) {
	var user core.User

	err := s.db.View().Where("user_id = ?", mixinUserID).First(&user).Error
	if store.IsErrNotFound(err) {
		return &core.User{}, nil
	}
	return &user, nil
}

func (s *userStore) FindByAddress(ctx context.Context, address string) (*core.User, error) {
	var user core.User
	err := s.db.View().Where("address = ?", address).First(&user).Error
	if store.IsErrNotFound(err) {
		// fail back to v0 address
		if err = s.db.View().Where("address_v0 = ?", address).First(&user).Error; store.IsErrNotFound(err) {
			return &core.User{}, nil
		}
	}
	return &user, err
}
