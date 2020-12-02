package core

import (
	"context"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

var (
	// GasCost gas cost
	GasCost = decimal.NewFromFloat(0.00000001)
)

// Wallet wallet
type Wallet struct {
	Client *mixin.Client `json:"client"`
	Pin    string        `json:"pin"`
}

// Snapshot snapshot
type Snapshot struct {
	ID         uint64          `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
	SnapshotID string          `sql:"size:36;unique_index:snapshot_idx" json:"snapshot_id,omitempty"`
	TraceID    string          `sql:"size:36;unique_index:trace_idx" json:"trace_id,omitempty"`
	UserID     string          `sql:"size:36" json:"user_id,omitempty"`
	OpponentID string          `sql:"size:36" json:"opponent_id,omitempty"`
	AssetID    string          `sql:"size:36" json:"asset_id,omitempty"`
	Amount     decimal.Decimal `sql:"type:decimal(20,8)" json:"amount,omitempty"`
	Memo       string          `sql:"size:256" json:"memo,omitempty"`
	CreatedAt  time.Time       `json:"created_at,omitempty"`
}

// ISnapshotStore snapshot store interface
type ISnapshotStore interface {
	Save(ctx context.Context, snapshot *Snapshot) error
	Find(ctx context.Context, snapshotID string) (*Snapshot, error)
	DeleteByTime(t time.Time) error
}

// IWalletService wallet service interface
type IWalletService interface {
	HandleTransfer(ctx context.Context, transfer *Transfer) (*Snapshot, error)
	PullSnapshots(ctx context.Context, cursor string, limit int) ([]*Snapshot, string, error)
	NewWallet(ctx context.Context, walletName, pin string) (*mixin.Keystore, string, error)
	PaySchemaURL(amount decimal.Decimal, asset, recipient, trace, memo string) (string, error)
	VerifyPayment(ctx context.Context, input *mixin.TransferInput) bool
}
