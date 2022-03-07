package core

import (
	"context"
	"fmt"

	"github.com/fox-one/pkg/uuid"
)

// User user info
// binding user_id and address
type User struct {
	ID        uint64 `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
	UserID    string `sql:"size:36;" json:"user_id,omitempty"`
	Address   string `sql:"size:36;" json:"address,omitempty"`
	AddressV0 string `sql:"size:36;" json:"address_v0,omitempty"`
}

// BuildUserAddress build compound user address
func BuildUserAddressV0(mixinUserID string) string {
	return uuid.MD5(fmt.Sprintf("compound-%s", mixinUserID))
}

// UserStore user store interface
type UserStore interface {
	List(ctx context.Context, from uint64, limit int) ([]*User, error)
	Create(ctx context.Context, user *User) error
	MigrateToV1(ctx context.Context, users []*User) error
	Find(ctx context.Context, mixinUserID string) (*User, error)
	FindByAddress(ctx context.Context, address string) (*User, error)
}
