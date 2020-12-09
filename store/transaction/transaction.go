package transaction

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

type transactionStore struct {
	db *db.DB
}

// New new transaction store
func New(db *db.DB) core.TransactionStore {
	return &transactionStore{}
}

func (s *transactionStore) Create(ctx context.Context, transactions ...*core.Transaction) error {
	return nil
}
func (s *transactionStore) Update(ctx context.Context, transaction *core.Transaction) error {
	return nil
}
func (s *transactionStore) List(ctx context.Context, offset int, limit int) ([]*core.Transaction, error) {
	return nil, nil
}
