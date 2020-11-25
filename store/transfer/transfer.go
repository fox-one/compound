package transfer

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/store/db"
)

type transferStore struct {
	db *db.DB
}

// New new transfer store
func New(db *db.DB) core.ITransferStore {
	return &transferStore{
		db: db,
	}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Transfer{})
		if err := tx.AutoMigrate(core.Transfer{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *transferStore) Create(ctx context.Context, tx *db.DB, transfer *core.Transfer) error {
	return tx.Update().Where("trace_id=?", transfer.TraceID).FirstOrCreate(transfer).Error
}
func (s *transferStore) Delete(ctx context.Context, tx *db.DB, ids ...uint64) error {
	return tx.Update().Where("id in (?)", ids).Delete(core.Transfer{}).Error
}
func (s *transferStore) Top(ctx context.Context, limit int) ([]*core.Transfer, error) {
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}

	var transfers []*core.Transfer
	if e := s.db.View().Limit(limit).Offset(0).Order("created_at ASC").Find(&transfers).Error; e != nil {
		return nil, e
	}

	return transfers, nil
}
