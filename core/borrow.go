package core

import (
	"context"
	"errors"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// Borrow borrow info
type Borrow struct {
	UserID    string          `sql:"size:36;PRIMARY_KEY" json:"user_id"`
	Symbol    string          `sql:"size:20;PRIMARY_KEY" json:"symbol"`
	CTokens   decimal.Decimal `sql:"type:decimal(20,8)" json:"c_tokens"`
	Interest  decimal.Decimal `sql:"type:decimal(20,8)" json:"interest"`
	Version   uint64          `sql:"default:0" json:"version"`
	CreatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

var (
	// ErrCollateralDisabled collateral disabled
	ErrCollateralDisabled = errors.New("collateral disabled")
	// ErrBorrowsOverCap borrows over cap
	ErrBorrowsOverCap = errors.New("borrows over market borrow cap")
)

// IBorrowStore supply store interface
type IBorrowStore interface {
	Save(ctx context.Context, borrow *Borrow) error
	Find(ctx context.Context, userID string, symbol string) (*Borrow, error)
	FindByUser(ctx context.Context, userID string) ([]*Borrow, error)
	Update(ctx context.Context, tx *db.DB, borrow *Borrow) error
}

// IBorrowService supply service interface
type IBorrowService interface {
	Repay(ctx context.Context, amount decimal.Decimal, userID string, market *Market) error
	Borrow(ctx context.Context, amount decimal.Decimal, userID string, market *Market) error
}

// SeizeTokens()
// seizeAllowed()
//

// RepayBorrow()
// borrowVerify()
// repayAllowed()
//LiquidateBorrow()
