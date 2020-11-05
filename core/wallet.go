package core

import (
	"context"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

// Wallet wallet
type Wallet struct {
	Client *mixin.Client `json:"client"`
	Pin    string        `json:"pin"`
}

// Transfer transfer record
type Transfer struct {
	ID         int64           `sql:"PRIMARY_KEY" json:"id,omitempty"`
	CreatedAt  time.Time       `json:"created_at,omitempty"`
	TraceID    string          `sql:"size:36" json:"trace_id,omitempty"`
	OpponentID string          `sql:"size:36" json:"opponent_id,omitempty"`
	AssetID    string          `sql:"size:36" json:"asset_id,omitempty"`
	Amount     decimal.Decimal `sql:"type:varchar(24)" json:"amount,omitempty"`
	Memo       string          `sql:"size:140" json:"memo,omitempty"`
}

// Snapshot snapshot
type Snapshot struct {
	ID         string          `json:"id,omitempty"`
	TraceID    string          `json:"trace_id,omitempty"`
	CreatedAt  time.Time       `json:"created_at,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	OpponentID string          `json:"opponent_id,omitempty"`
	AssetID    string          `json:"asset_id,omitempty"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	Memo       string          `json:"memo,omitempty"`
}

// IWalletStore wallet store interface
type IWalletStore interface {
	CreateTransfer(ctx context.Context, transfer *Transfer) error
	DeleteTransfers(ctx context.Context, transfers []*Transfer) error
	ListTransfers(ctx context.Context, limit int) ([]*Transfer, error)
}

// IWalletService wallet service interface
type IWalletService interface {
	HandleTransfer(ctx context.Context, transfer *Transfer) (*Snapshot, error)
	PullSnapshots(ctx context.Context, cursor string, limit int) ([]*Snapshot, string, error)
	NewWallet(ctx context.Context, walletName, pin string) (*mixin.Keystore, string, error)
	PaySchemaURL(amount decimal.Decimal, asset, recipient, trace, memo string) (string, error)
	VerifyPayment(ctx context.Context, input *mixin.TransferInput) bool
}
