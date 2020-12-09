package core

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

// Transaction transaction ifo
type Transaction struct {
	ID        int64           `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id,omitempty"`
	Type      ActionType      `json:"type,omitempty"`
	TraceID   string          `sql:"size:36;unique_index:idx_transactions_trace_id" json:"trace_id,omitempty"`
	UserID    string          `sql:"size:36;index:idx_transactions_user_id" json:"user_id,omitempty"`
	FollowID  string          `sql:"size:36;index:idx_transactions_follow_id" json:"follow_id,omitempty"`
	AssetID   string          `sql:"size:36;index:idx_transactions_asset_id" json:"asset_id,omitempty"`
	Amount    decimal.Decimal `sql:"type:decimal(28,8)" json:"amount,omitempty"`
	Data      types.JSONText  `sql:"type:TEXT" json:"data,omitempty"`
	CreatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	UpdatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
}

// TransactionStore transaction store interface
type TransactionStore interface {
	Create(ctx context.Context, transactions ...*Transaction) error
	Update(ctx context.Context, transaction *Transaction) error
	List(ctx context.Context, offset int, limit int) ([]*Transaction, error)
}
