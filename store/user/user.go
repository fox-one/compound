package user

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

type userStore struct {
	db *db.DB
}

func New(db *db.DB) core.IUserStore {
	return &userStore{db: db}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.User{})
		if err := tx.AutoMigrate(core.User{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func toUpdateParams(user *core.User) map[string]interface{} {
	return map[string]interface{}{
		"name":         user.Name,
		"avatar":       user.Avatar,
		"access_token": user.AccessToken,
	}
}

func (s *userStore) update(ctx context.Context, user *core.User) error {
	updates := toUpdateParams(user)
	updates["version"] = user.Version + 1

	tx := s.db.Update().Model(user).Where("version = ?", user.Version).Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return db.ErrOptimisticLock
	}

	return nil
}

// Save insert or update user
func (s *userStore) Save(ctx context.Context, user *core.User) error {
	if user.ID == 0 {
		return s.db.Update().Create(user).Error
	}

	return s.update(ctx, user)
}

func (s *userStore) Find(ctx context.Context, mixinID string) (*core.User, error) {
	var user core.User
	if err := s.db.View().Where("mixin_id=?", mixinID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userStore) All(ctx context.Context) ([]*core.User, error) {
	var users []*core.User
	if err := s.db.View().Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
