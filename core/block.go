package core

import "context"

// BlockMemo 每个时间属于一个块，同一个时间块，traceID一致，一个块在链上至多有且只有一个transaction, 默认15s一个块
type BlockMemo struct {
	Service string `json:"s"`
	Block   int64  `json:"b"`
}

// IBlockService block service interface
type IBlockService interface {
	CurrentBlock(ctx context.Context) (int64, error)
	NewBlockMemo(ctx context.Context, currentBlock int64) (string, error)
	ParseBlockMemo(ctx context.Context, memoStr string) (*BlockMemo, error)
}
