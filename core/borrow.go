package core

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// Borrow user borrow model
type Borrow struct {
	ID            uint64          `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
	UserID        string          `sql:"size:36;unique_index:borrow_idx" json:"-"`
	AssetID       string          `sql:"size:36;unique_index:borrow_idx" json:"asset_id"`
	Principal     decimal.Decimal `sql:"type:decimal(32,16)" json:"principal"`
	InterestIndex decimal.Decimal `sql:"type:decimal(32,16);default:1" json:"interest_index"`
	Version       int64           `sql:"default:0" json:"version"`
	CreatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

var (
	// ErrCollateralDisabled collateral disabled
	ErrCollateralDisabled = errors.New("collateral disabled")
	// ErrBorrowsOverCap borrows over cap
	ErrBorrowsOverCap = errors.New("borrows over market borrow cap")
)

// IBorrowStore supply store interface
type IBorrowStore interface {
	Create(ctx context.Context, borrow *Borrow) error
	Find(ctx context.Context, userID string, assetID string) (*Borrow, error)
	FindByUser(ctx context.Context, userID string) ([]*Borrow, error)
	FindByAssetID(ctx context.Context, assetID string) ([]*Borrow, error)
	CountOfBorrowers(ctx context.Context, assetID string) (int64, error)
	Update(ctx context.Context, borrow *Borrow, version int64) error
	All(ctx context.Context) ([]*Borrow, error)
	Users(ctx context.Context) ([]string, error)
}
