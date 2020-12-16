package core

import "context"

// User user info
type User struct {
	ID     uint64 `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
	UserID string `sql:"size:36;unique_index:idx_users" json:"user_id"`
}

// UserStore user store interface
type UserStore interface {
	Save(ctx context.Context, user *User) error
	Find(ctx context.Context, mixinUserID string) (*User, error)
}
