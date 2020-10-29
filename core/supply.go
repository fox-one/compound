package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Supply supply info
type Supply struct {
	ID        uint64          `sql:"AUTO_INCREMENT;PRIMARY_KEY" json:"id"`
	UserID    string          `sql:"size:36;unique_index:user_symbol_idx" json:"user_id"`
	Symbol    string          `sql:"size:20;unique_index:user_symbol_idx" json:"symbol"`
	CTokens   decimal.Decimal `sql:"type:decimal(20,8)" json:"c_tokens"`
	CreatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// ISupplyStore supply store interface
type ISupplyStore interface {
	Save(ctx context.Context, supply *Supply) error
	Find(ctx context.Context, userID string, symbols ...string) ([]*Supply, error)
}

// ISupplyService supply service interface
type ISupplyService interface {
	Redeem(ctx context.Context, amount decimal.Decimal, userID, symbol string) error
	Loan(ctx context.Context, amount decimal.Decimal, userID, symbol string) error
}
