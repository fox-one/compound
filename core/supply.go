package core

import (
	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// Supply supply info
type Supply struct {
	UserID    string          `sql:"size:36;PRIMARY_KEY" json:"user_id"`
	Symbol    string          `sql:"size:20;PRIMARY_KEY" json:"symbol"`
	CTokens   decimal.Decimal `sql:"type:decimal(20,8)" json:"c_tokens"`
	Version   uint64          `sql:"default:0" json:"version"`
	CreatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// ISupplyStore supply store interface
type ISupplyStore interface {
	Save(ctx context.Context, supply *Supply) error
	Find(ctx context.Context, userID string, symbols ...string) ([]*Supply, error)
	Update(ctx context.Context, tx *db.DB, supply *Supply) error
}

// ISupplyService supply service interface
type ISupplyService interface {
	Redeem(ctx context.Context, amount decimal.Decimal, market *Market) (string, error)
	Loan(ctx context.Context, amount decimal.Decimal, market *Market) (string, error)
}
