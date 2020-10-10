package core

import (
	"context"

	"github.com/shopspring/decimal"
)

// Asset asset struct
type Asset struct {
	ID      string          `sql:"size:36;PRIMARY_KEY" json:"id"`
	Name    string          `sql:"size:64" json:"name,omitempty"`
	Symbol  string          `sql:"size:32" json:"symbol,omitempty"`
	Logo    string          `sql:"size:256" json:"logo,omitempty"`
	ChainID string          `sql:"size:36" json:"chain_id,omitempty"`
	Price   decimal.Decimal `sql:"type:decimal(24,8)" json:"price,omitempty"`
}

// IAssetStore asset store interface
type IAssetStore interface {
	Save(ctx context.Context, asset *Asset) error
	Find(ctx context.Context, id string) (*Asset, error)
	All(ctx context.Context) ([]*Asset, error)
}

// IAssetService asset service interface
type IAssetService interface {
	Find(ctx context.Context, id string) (*Asset, error)
	All(ctx context.Context) ([]*Asset, error)
}
