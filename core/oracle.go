package core

import (
	"context"
	"time"
)

type OracleSigner struct {
	ID        int64     `sql:"PRIMARY_KEY" json:"id,omitempty"`
	UserID    string    `sql:"size:36;unique_index:idx_oracle_signers_user_id" json:"user_id,omitempty"`
	PublicKey string    `sql:"size:256" json:"public_key,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type OracleSignerStore interface {
	Save(ctx context.Context, userID, publicKey string) error
	Delete(ctx context.Context, userID string) error
	FindAll(ctx context.Context) ([]*OracleSigner, error)
}
