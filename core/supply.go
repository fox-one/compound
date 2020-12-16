package core

import (
	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// Supply supply info
type Supply struct {
	ID            uint64          `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
	UserID        string          `sql:"size:36;unique_index:supply_idx" json:"-"`
	CTokenAssetID string          `sql:"size:36;unique_index:supply_idx" json:"ctoken_asset_id"`
	Collaterals   decimal.Decimal `sql:"type:decimal(32,16)" json:"collaterals"`
	Version       int64           `sql:"default:0" json:"version"`
	CreatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// ISupplyStore supply store interface
type ISupplyStore interface {
	Save(ctx context.Context, tx *db.DB, supply *Supply) error
	Find(ctx context.Context, userID string, ctokenAssetID string) (*Supply, bool, error)
	FindByUser(ctx context.Context, userID string) ([]*Supply, error)
	FindByCTokenAssetID(ctx context.Context, assetID string) ([]*Supply, error)
	SumOfSupplies(ctx context.Context, ctokenAssetID string) (decimal.Decimal, error)
	CountOfSuppliers(ctx context.Context, ctokenAssetID string) (int64, error)
	Update(ctx context.Context, tx *db.DB, supply *Supply) error
	All(ctx context.Context) ([]*Supply, error)
	Users(ctx context.Context) ([]string, error)
}

// ISupplyService supply service interface
type ISupplyService interface {
	RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, market *Market) bool
}
