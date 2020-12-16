package core

import (
	"compound/pkg/id"
	"fmt"

	"context"
)

// User user info
type User struct {
	ID      uint64 `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
	UserID  string `sql:"size:36;unique_index:idx_users_user_id" json:"user_id"`
	Address string `sql:"size:36;unique_index:idx_users_address" json:"address"`
}

// BuildUserAddress build compound user address
func BuildUserAddress(mixinUserID string) string {
	return id.UUIDFromString(fmt.Sprintf("compound-%s", mixinUserID))
}

// UserStore user store interface
type UserStore interface {
	Save(ctx context.Context, user *User) error
	Find(ctx context.Context, mixinUserID string) (*User, error)
	FindByAddress(ctx context.Context, address string) (*User, error)
}
