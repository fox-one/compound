package request

import (
	"context"

	"compound/core"
)

type key int

const (
	userKey key = iota
)

type ContextX struct {
	context.Context
}

// NewContext context extension
func NewContext(ctx context.Context) ContextX {
	return ContextX{
		Context: ctx,
	}
}

// WithUser context with user
func (c ContextX) WithUser(user *core.User) context.Context {
	return context.WithValue(c, userKey, user)
}

// GetUser get user from context
func (c ContextX) GetUser() (*core.User, bool) {
	user, ok := c.Value(userKey).(*core.User)
	return user, ok
}
