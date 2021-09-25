package sysversion

import (
	"context"

	"github.com/fox-one/pkg/property"
)

const (
	SysVersionKey = "sysversion"
)

func ReadSysVersion(ctx context.Context, property property.Store) (int64, error) {
	v, err := property.Get(ctx, SysVersionKey)
	if err != nil {
		return 0, err
	}
	return v.Int64(), nil
}
