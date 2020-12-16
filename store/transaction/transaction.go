package transaction

import (
	"compound/core"
	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
)

type transactionStore struct {
	db *db.DB
}

// New new transaction store
func New(db *db.DB) core.TransactionStore {
	return &transactionStore{
		db: db,
	}
}

func (s *transactionStore) Create(ctx context.Context, tx *db.DB, transaction *core.Transaction) error {
	return tx.Update().Where("trace_id=?", transaction.TraceID).FirstOrCreate(transaction).Error
}

func (s *transactionStore) Update(ctx context.Context, tx *db.DB, transaction *core.Transaction) error {
	return tx.Update().Model(core.Transaction{}).Where("trace_id=?", transaction.TraceID).Updates(transaction).Error
}

func (s *transactionStore) List(ctx context.Context, offset time.Time, limit int) ([]*core.Transaction, error) {
	var transactions []*core.Transaction
	if limit <= 0 {
		limit = 500
	}

	if err := s.db.View().Where("created_at >=?", offset).Order("created_at ASC").Limit(limit).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}
