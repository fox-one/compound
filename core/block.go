package core

import (
	"context"
	"time"
)

// IBlockService block service interface
// 15 seconds per block
type IBlockService interface {
	// get block number by time
	GetBlock(ctx context.Context, t time.Time) (int64, error)
}
