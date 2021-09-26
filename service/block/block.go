package block

import (
	"compound/core"
	"compound/internal/compound"
	"context"
	"time"
)

type (
	Config struct {
		Genesis int64
	}

	service struct {
		genesis int64
	}
)

// New new block service
func New(config Config) core.IBlockService {
	return &service{
		genesis: config.Genesis,
	}
}

// GetBlock get block by time
func (s *service) GetBlock(ctx context.Context, t time.Time) (int64, error) {
	block, e := compound.GetBlockByTime(ctx, compound.SecondsPerBlock, s.genesis, t)
	if e != nil {
		return 0, e
	}
	return block, nil
}
