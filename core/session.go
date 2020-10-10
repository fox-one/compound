package core

import (
	"context"
)

// Session user session
type Session interface {
	// Login return user mixin id
	Login(ctx context.Context, accessToken string) (*User, error)
}
