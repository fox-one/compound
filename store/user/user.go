package user

import (
	"compound/core"
	"context"

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

func (s *userStore) Save(ctx context.Context, user *core.User) error {
	return s.db.Update().Where("user_id=?", user.UserID).FirstOrCreate(user).Error
}

func (s *userStore) Find(ctx context.Context, mixinUserID string) (*core.User, error) {
	var user core.User
	if err := s.db.View().Where("user_id=?", mixinUserID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
