package core

import (
	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// Transfer transfer struct
type Transfer struct {
	ID         uint64          `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id,omitempty"`
	TraceID    string          `sql:"size:36;unique_index:trace_idx" json:"trace_id,omitempty"`
	OpponentID string          `sql:"size:36" json:"opponent_id,omitempty"`
	AssetID    string          `sql:"size:36" json:"asset_id,omitempty"`
	Amount     decimal.Decimal `sql:"type:decimal(20,8)" json:"amount,omitempty"`
	Memo       string          `sql:"size:140" json:"memo,omitempty"`
	Status     string          `sql:"size:20" json:"status,omitempty"`
	Version    int64           `sql:"default:0" json:"version,omitempty"`
	CreatedAt  time.Time       `json:"created_at,omitempty"`
}

const (
	// TransferStatusPending pending
	TransferStatusPending = "pending"
	// TransferStatusDone done
	TransferStatusDone = "done"
)

// ITransferStore transfer store interface
type ITransferStore interface {
	Create(ctx context.Context, tx *db.DB, transfer *Transfer) error
	FindByStatus(ctx context.Context, status string) ([]*Transfer, error)
	UpdateStatus(ctx context.Context, tx *db.DB, id uint64, status string) error
	DeleteByTime(t time.Time) error
}
