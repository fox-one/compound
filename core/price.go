package core

import (
	"context"
	"database/sql"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

// Price price info
type Price struct {
	ID          int64           `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id,omitempty"`
	AssetID     string          `sql:"size:36;unique_index:idx_prices" json:"asset_id,omitempty"`
	BlockNumber int64           `sql:"default:0;unique_index:idx_prices" json:"block_number,omitempty"`
	Price       decimal.Decimal `sql:"type:decimal(20,8)" json:"price,omitempty"`
	Content     types.JSONText  `sql:"type:varchar(1024)" json:"content,omitempty"`
	Version     int64           `sql:"default:0" json:"version,omitempty"`
	PassedAt    sql.NullTime    `json:"passed_at,omitempty"`
	CreatedAt   time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	UpdatedAt   time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
}

// PriceTicker price ticker
type PriceTicker struct {
	Provider string          `json:"provider,omitempty"`
	Symbol   string          `json:"symbol,omitempty"`
	Price    decimal.Decimal `json:"price,omitempty"`
}

// IPriceStore price store interface
type IPriceStore interface {
	Create(ctx context.Context, tx *db.DB, price *Price) error
	FindByAssetBlock(ctx context.Context, assetID string, blockNumber int64) (*Price, bool, error)
	Update(ctx context.Context, tx *db.DB, price *Price) error
}

// IPriceOracleService pracle price service interface
type IPriceOracleService interface {
	GetCurrentUnderlyingPrice(ctx context.Context, market *Market) (decimal.Decimal, error)
	PullPriceTicker(ctx context.Context, symbol string, t time.Time) (*PriceTicker, error)
	PullAllPriceTickers(ctx context.Context, t time.Time) ([]*PriceTicker, error)
}
