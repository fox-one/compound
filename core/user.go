package core

import (
	"context"
	"time"
)

// User user model
type User struct {
	ID          int64     `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id,omitempty"`
	CreatedAt   time.Time `sql:"default:CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	UpdatedAt   time.Time `sql:"default:CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
	Version     int64     `json:"version,omitempty"`
	MixinID     string    `sql:"size:36;UNIQUE_INDEX:idx_mixin_id" json:"mixin_id,omitempty"`
	Role        string    `sql:"size:24" json:"role,omitempty"`
	Lang        string    `sql:"size:36" json:"lang,omitempty"`
	Name        string    `sql:"size:64" json:"name,omitempty"`
	Avatar      string    `sql:"size:255" json:"avatar,omitempty"`
	AccessToken string    `sql:"size:512" json:"access_token,omitempty"`
}

// IUserStore user store interface
type IUserStore interface {
	Save(ctx context.Context, user *User) error
	Find(ctx context.Context, mixinID string) (*User, error)
	All(ctx context.Context) ([]*User, error)
}

// IUserService user service interface
type IUserService interface {
	Find(ctx context.Context, mixinID string) (*User, error)
	Login(ctx context.Context, token string) (*User, error)
	SupportMarket(ctx context.Context)
}
