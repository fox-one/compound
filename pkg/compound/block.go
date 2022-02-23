package compound

import (
	"context"
	"errors"
	"time"
)

var (
	genesis int64 = 0
)

func SetupGenesis(_genesis int64) {
	genesis = _genesis
}

// GetBlockByTime get block by time
func GetBlockByTime(ctx context.Context, t time.Time) (int64, error) {
	seconds := t.UTC().Unix() - genesis
	if seconds <= 0 {
		return 0, errors.New("invalid blocks")
	}

	return seconds / SecondsPerBlock, nil
}
