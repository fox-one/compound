package core

import (
	"context"
	"time"
)

// BlockMemoKey block memo key
type BlockMemoKey string

const (
	// BlockMemoKeyService key service type :string
	BlockMemoKeyService BlockMemoKey = "srv"
	// BlockMemoKeyBlock block index :int64
	BlockMemoKeyBlock BlockMemoKey = "b"
	// BlockMemoKeySymbol symbol key :string
	BlockMemoKeySymbol BlockMemoKey = "sb"
	// BlockMemoKeyPrice price :decimal
	BlockMemoKeyPrice BlockMemoKey = "pr"
	// BlockMemoKeyUtilizationRate utilization rate :decimal
	BlockMemoKeyUtilizationRate BlockMemoKey = "ur"
	// BlockMemoKeyBorrowRate borrow rate :decimal
	BlockMemoKeyBorrowRate BlockMemoKey = "br"
	// BlockMemoKeySupplyRate supply rate : decimal
	BlockMemoKeySupplyRate BlockMemoKey = "sr"
)

func (k BlockMemoKey) String() string {
	return string(k)
}

const (
	// MemoServiceBlock block
	MemoServiceBlock = "blk"
	// MemoServicePrice prc
	MemoServicePrice = "prc"
	// MemoServiceMarket market
	MemoServiceMarket = "mkt"
)

// BlockMemo 每个时间属于一个块，同一个时间块，traceID一致，一个块在链上至多有且只有一个transaction, 默认15s一个块
type BlockMemo map[BlockMemoKey]string

// IBlockService block service interface
type IBlockService interface {
	GetBlock(ctx context.Context, t time.Time) (int64, error)
	CurrentBlock(ctx context.Context) (int64, error)
	FormatBlockMemo(ctx context.Context, memo BlockMemo) (string, error)
	ParseBlockMemo(ctx context.Context, memoStr string) (BlockMemo, error)
}
