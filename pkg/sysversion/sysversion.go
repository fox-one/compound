package sysversion

import (
	"context"

	"github.com/fox-one/pkg/property"
)

const (
	SysVersionKey = "sysversion"
)

type (
	sysverKey struct{}
)

func ReadSysVersion(ctx context.Context, property property.Store) (int64, error) {
	v, err := property.Get(ctx, SysVersionKey)
	if err != nil {
		return 0, err
	}
	return v.Int64(), nil
}

func WithContext(ctx context.Context, syversion int64) context.Context {
	return context.WithValue(ctx, sysverKey{}, syversion)
}

func FromContext(ctx context.Context) int64 {
	if v := ctx.Value(sysverKey{}); v != nil {
		if sysver, ok := v.(int64); ok {
			return sysver
		}
	}
	return 0
}
