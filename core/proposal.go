package core

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type (
	// Proposal proposal info
	Proposal struct {
		ID        int64           `sql:"PRIMARY_KEY" json:"id,omitempty"`
		CreatedAt time.Time       `json:"created_at,omitempty"`
		UpdatedAt time.Time       `json:"updated_at,omitempty"`
		PassedAt  sql.NullTime    `json:"passed_at,omitempty"`
		Version   int64           `json:"version,omitempty"`
		TraceID   string          `sql:"size:36" json:"trace_id,omitempty"`
		Creator   string          `sql:"size:36" json:"creator,omitempty"`
		AssetID   string          `sql:"size:36" json:"asset_id,omitempty"`
		Amount    decimal.Decimal `sql:"type:decimal(32,8)" json:"amount,omitempty"`
		Action    ActionType      `json:"action,omitempty"`
		Content   types.JSONText  `sql:"type:varchar(1024)" json:"content,omitempty"`
		Votes     pq.StringArray  `sql:"type:varchar(1024)" json:"votes,omitempty"`
	}

	ProposalItem struct {
		// Key is the parameter name
		Key string `json:"key,omitempty"`
		// Value the proposal applied
		Value string `json:"value,omitempty"`
		// Hint the parameter hint
		Hint string `json:"hint,omitempty"`
		// Action the value applied
		Action string `json:"action,omitempty"`
	}

	// ProposalStore proposal store interface
	ProposalStore interface {
		Create(ctx context.Context, proposal *Proposal) error
		Find(ctx context.Context, trace string) (*Proposal, bool, error)
		Update(ctx context.Context, proposal *Proposal, version int64) error
		List(ctx context.Context, fromID int64, limit int) ([]*Proposal, error)
	}

	// ProposalService proposal service interface
	ProposalService interface {
		ListItems(ctx context.Context, proposal *Proposal) ([]ProposalItem, error)
		ProposalCreated(ctx context.Context, proposal *Proposal, by string, sysver int64) error
		ProposalApproved(ctx context.Context, proposal *Proposal, by string, sysver int64) error
		ProposalPassed(ctx context.Context, proposal *Proposal, sysver int64) error
	}
)
