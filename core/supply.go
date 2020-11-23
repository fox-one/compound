package core

import (
	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// Supply supply info
type Supply struct {
	UserID        string          `sql:"size:36;PRIMARY_KEY" json:"user_id"`
	Symbol        string          `sql:"size:20;PRIMARY_KEY" json:"symbol"`
	CTokenAssetID string          `sql:"size:36;PRIMARY_KEY" json:"ctoken_asset_id"`
	Collaterals   decimal.Decimal `sql:"type:decimal(20,8)" json:"collaterals"`
	Version       int64           `sql:"default:0" json:"version"`
	CreatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// ISupplyStore supply store interface
type ISupplyStore interface {
	Save(ctx context.Context, tx *db.DB, supply *Supply) error
	Find(ctx context.Context, userID string, ctokenAssetID string) (*Supply, error)
	FindByUser(ctx context.Context, userID string) ([]*Supply, error)
	FindByCTokenAssetID(ctx context.Context, assetID string) ([]*Supply, error)
	SumOfSupplies(ctx context.Context, symbol string) (decimal.Decimal, error)
	CountOfSuppliers(ctx context.Context, symbol string) (int64, error)
	Update(ctx context.Context, tx *db.DB, supply *Supply) error
	All(ctx context.Context) ([]*Supply, error)
	Users(ctx context.Context) ([]string, error)
}

// ISupplyService supply service interface
type ISupplyService interface {
	RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, market *Market) bool
}
