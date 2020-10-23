package block

import (
	"compound/core"
	"context"
	"errors"
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

//CurrentBlock 获取当前块高度
func (s *service) CurrentBlock(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	seconds := now.Unix()

	if s.config.App.SecondsPerBlock <= 0 {
		return 0, errors.New("invalid seconds_per_block")
	}
	currentBlock := seconds / s.config.App.SecondsPerBlock

	return currentBlock, nil
}

func (s *service) NewBlockMemo(ctx context.Context, currentBlock int) (string, error) {
	return "", nil
}
