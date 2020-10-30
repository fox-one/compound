package borrow

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

type borrowStore struct {
	db *db.DB
}

// New new borrow store
func New(db *db.DB) core.IBorrowStore {
	return &borrowStore{
		db: db,
	}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Borrow{})
		if err := tx.AutoMigrate(core.Borrow{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *borrowStore) Save(ctx context.Context, borrow *core.Borrow) error {
	if e := s.db.Update().Where("user_id=? and symbol=?", borrow.UserID, borrow.Symbol).FirstOrCreate(borrow).Error; e != nil {
		return e
	}

	return nil
}
func (s *borrowStore) Find(ctx context.Context, userID string, symbols ...string) ([]*core.Borrow, error) {
	query := s.db.View().Where("user_id=?", userID)
	if len(symbols) > 0 {
		query = query.Where("symbol in (?)", symbols)
	}
	var borrows []*core.Borrow
	if e := query.Find(&borrows).Error; e != nil {
		return nil, e
	}

	return borrows, nil
}
func (s *borrowStore) Update(ctx context.Context, tx *db.DB, borrow *core.Borrow) error {
	version := borrow.Version
	borrow.Version++
	if err := tx.Update().Model(core.Supply{}).Where("user_id=? and symbol=? and version=?", borrow.UserID, borrow.Symbol, version).Updates(borrow).Error; err != nil {
		return err
	}

	return nil
}
