package block

import (
	"compound/core"
	"compound/internal/compound"
	"context"
	"time"
)

type service struct {
	config *core.Config
}

// New new block service
func New(config *core.Config) core.IBlockService {
	return &service{
		config: config,
	}
}

// GetBlock get block by time
func (s *service) GetBlock(ctx context.Context, t time.Time) (int64, error) {
	block, e := compound.GetBlockByTime(ctx, compound.SecondsPerBlock, s.config.Genesis, t)
	if e != nil {
		return 0, e
	}
	return block, nil
}
