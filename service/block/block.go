package block

import (
	"compound/core"
	"compound/internal/compound"
	"context"
	"encoding/json"
	"errors"
)

const (
	// ServiceType service type
	ServiceType = "block"
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
	current, e := compound.CurrentBlock(ctx, s.config.App.SecondsPerBlock, s.config.App.Genesis)
	if e != nil {
		return 0, e
	}
	return current, nil
}

func (s *service) NewBlockMemo(ctx context.Context, currentBlock int64) (string, error) {
	memo := core.BlockMemo{
		Service: ServiceType,
		Block:   currentBlock,
	}

	bs, e := json.Marshal(&memo)
	if e != nil {
		return "", e
	}

	return string(bs), nil
}

func (s *service) ParseBlockMemo(ctx context.Context, memoStr string) (*core.BlockMemo, error) {
	var memo core.BlockMemo
	e := json.Unmarshal([]byte(memoStr), &memo)
	if e != nil {
		return nil, e
	}

	if memo.Service != ServiceType {
		return nil, errors.New("invalid service type")
	}

	return &memo, nil
}
