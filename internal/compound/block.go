package compound

import (
	"context"
	"errors"
	"time"
)

// CurrentBlock current block
func CurrentBlock(ctx context.Context, secondsPerBlock, genesis int64) (int64, error) {
	return GetBlockByTime(ctx, secondsPerBlock, genesis, time.Now())
}

// GetBlockByTime get block by time
func GetBlockByTime(ctx context.Context, secondsPerBlock, genesis int64, t time.Time) (int64, error) {
	if secondsPerBlock <= 0 {
		return 0, errors.New("secondsPerBlock should not be less than or equal zero")
	}

	seconds := t.UTC().Unix() - genesis

	if seconds <= 0 {
		return 0, errors.New("invalid blocks")
	}

	currentBlock := seconds / secondsPerBlock

	return currentBlock, nil
}
