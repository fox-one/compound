package transaction

import (
	"compound/core"
	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
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

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Transaction{})

		if err := tx.AutoMigrate(core.Transaction{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *transactionStore) Create(ctx context.Context, transaction *core.Transaction) error {
	return s.db.Update().Where("trace_id=?", transaction.TraceID).FirstOrCreate(transaction).Error
}

func (s *transactionStore) FindByTraceID(ctx context.Context, traceID string) (*core.Transaction, error) {
	var transaction core.Transaction
	if err := s.db.View().Where("trace_id=?", traceID).First(&transaction).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &core.Transaction{}, nil
		}
		return nil, err
	}

	return &transaction, nil
}

func (s *transactionStore) Update(ctx context.Context, transaction *core.Transaction) error {
	return s.db.Update().Model(core.Transaction{}).Where("trace_id=?", transaction.TraceID).Updates(transaction).Error
}

func (s *transactionStore) List(ctx context.Context, offset time.Time, limit int, status core.TransactionStatus) ([]*core.Transaction, error) {
	var transactions []*core.Transaction
	if limit <= 0 {
		limit = 500
	}

	if err := s.db.View().Where("status=? and created_at >=?", status, offset).Order("created_at ASC").Limit(limit).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}
