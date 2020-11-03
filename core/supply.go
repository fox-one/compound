package core

import (
	"context"
	"strconv"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// CollateStatus 抵押状态
type CollateStatus int

const (
	// CollateStateOff off
	CollateStateOff CollateStatus = iota
	// CollateStateOn on
	CollateStateOn
)

func (s CollateStatus) String() string {
	return strconv.Itoa(int(s))
}

// Supply supply info
type Supply struct {
	UserID        string          `sql:"size:36;PRIMARY_KEY" json:"user_id"`
	Symbol        string          `sql:"size:20;PRIMARY_KEY" json:"symbol"`
	Principal     decimal.Decimal `sql:"type:decimal(20,8)" json:"principal"`
	InterestIndex decimal.Decimal `sql:"type:decimal(20,8)" json:"interest_index"`
	Ctokens       decimal.Decimal `sql:"type:decimal(20,8)" json:"ctokens"`
	//是否可抵押
	CollateStatus CollateStatus `sql:"default:0" json:"collate_status"`
	Version       uint64        `sql:"default:0" json:"version"`
	CreatedAt     time.Time     `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time     `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
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
	Redeem(ctx context.Context, amount decimal.Decimal, userID string, market *Market) (string, error)
	Supply(ctx context.Context, amount decimal.Decimal, market *Market) (string, error)
	SetCollateStatus(ctx context.Context, userID string, market *Market, status CollateStatus) error
}
