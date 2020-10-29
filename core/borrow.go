package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Borrow borrow info
type Borrow struct {
	ID        uint64          `sql:"AUTO_INCREMENT;PRIMARY_KEY" json:"id"`
	UserID    string          `sql:"size:36;unique_index:user_symbol_idx" json:"user_id"`
	Symbol    string          `sql:"size:20;unique_index:user_symbol_idx" json:"symbol"`
	CTokens   decimal.Decimal `sql:"type:decimal(20,8)" json:"c_tokens"`
	CreatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// IBorrowStore supply store interface
type IBorrowStore interface {
	Save(ctx context.Context, borrow *Borrow) error
	Find(ctx context.Context, userID string, symbols ...string) ([]*Borrow, error)
}

// IBorrowService supply service interface
type IBorrowService interface {
	Repay(ctx context.Context, amount decimal.Decimal, userID, symbol string) error
	Borrow(ctx context.Context, amount decimal.Decimal, userID, symbol string) error
	
}

// SeizeTokens()
// seizeAllowed()
//

// RepayBorrow()
// borrowVerify()
// repayAllowed()
//LiquidateBorrow()
