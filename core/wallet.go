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
	ID         string          `json:"id,omitempty"`
	TraceID    string          `json:"trace_id,omitempty"`
	CreatedAt  time.Time       `json:"created_at,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	OpponentID string          `json:"opponent_id,omitempty"`
	AssetID    string          `json:"asset_id,omitempty"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	Memo       string          `json:"memo,omitempty"`
}

// IWalletService wallet service interface
type IWalletService interface {
	HandleTransfer(ctx context.Context, transfer *Transfer) (*Snapshot, error)
	PullSnapshots(ctx context.Context, cursor string, limit int) ([]*Snapshot, string, error)
	NewWallet(ctx context.Context, walletName, pin string) (*mixin.Keystore, string, error)
	PaySchemaURL(amount decimal.Decimal, asset, recipient, trace, memo string) (string, error)
	VerifyPayment(ctx context.Context, input *mixin.TransferInput) bool
}
