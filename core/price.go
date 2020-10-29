package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// PriceTicker price ticker
type PriceTicker struct {
	Provider string          `json:"provider"`
	Symbol   string          `json:"symbol"`
	Price    decimal.Decimal `json:"price"`
}

// IPriceOracleService pracle price service interface
type IPriceOracleService interface {
	Save(ctx context.Context, symbol string, price decimal.Decimal, block int64) error
	GetUnderlyingPrice(ctx context.Context, symbol string, block int64) (decimal.Decimal, error)
	PullPriceTicker(ctx context.Context, symbol string, t time.Time) (*PriceTicker, error)
	PullAllPriceTickers(ctx context.Context, t time.Time) ([]*PriceTicker, error)
}
