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
	Principal     decimal.Decimal `sql:"type:decimal(20,8)" json:"principal"`
	InterestIndex decimal.Decimal `sql:"type:decimal(20,8)" json:"interest_index"`
	// 总供应量凭证
	Ctokens decimal.Decimal `sql:"type:decimal(20,8)" json:"ctokens"`
	// 抵押量
	CollateTokens decimal.Decimal `sql:"type:decimal(20,8)" json:"collate_tokens"`
	Version       uint64          `sql:"default:0" json:"version"`
	CreatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// ISupplyStore supply store interface
type ISupplyStore interface {
	Save(ctx context.Context, supply *Supply) error
	Find(ctx context.Context, userID string, symbol string) (*Supply, error)
	FindByUser(ctx context.Context, userID string) ([]*Supply, error)
	Update(ctx context.Context, tx *db.DB, supply *Supply) error
}

// ISupplyService supply service interface
type ISupplyService interface {
	Redeem(ctx context.Context, redeemTokens decimal.Decimal, userID string, market *Market) (string, error)
	RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, userID string, market *Market) bool
	MaxRedeem(ctx context.Context, userID string, market *Market) (decimal.Decimal, error)
	Supply(ctx context.Context, amount decimal.Decimal, market *Market) (string, error)
	Pledge(ctx context.Context, pledgedTokens decimal.Decimal, userID string, market *Market) (string, error)
	Unpledge(ctx context.Context, pledgedTokens decimal.Decimal, userID string, market *Market) error
	MaxPledge(ctx context.Context, userID string, market *Market) (decimal.Decimal, error)
}
