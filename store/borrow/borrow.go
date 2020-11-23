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

func (s *borrowStore) Save(ctx context.Context, tx *db.DB, borrow *core.Borrow) error {
	if e := tx.Update().Where("user_id=? and symbol=?", borrow.UserID, borrow.Symbol).FirstOrCreate(borrow).Error; e != nil {
		return e
	}

	return nil
}

func (s *borrowStore) Find(ctx context.Context, userID string, symbol string) (*core.Borrow, error) {
	var borrow core.Borrow
	if e := s.db.View().Where("user_id=? and symbol=?", userID, symbol).First(&borrow).Error; e != nil {
		return nil, e
	}

	return &borrow, nil
}

func (s *borrowStore) FindByUser(ctx context.Context, userID string) ([]*core.Borrow, error) {
	var borrows []*core.Borrow
	if e := s.db.View().Where("user_id=?", userID).Find(&borrows).Error; e != nil {
		return nil, e
	}

	return borrows, nil
}

func (s *borrowStore) FindBySymbol(ctx context.Context, symbol string) ([]*core.Borrow, error) {
	var borrows []*core.Borrow
	if e := s.db.View().Where("symbol=?", symbol).Find(&borrows).Error; e != nil {
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

func (s *borrowStore) All(ctx context.Context) ([]*core.Borrow, error) {
	var borrows []*core.Borrow
	if e := s.db.View().Find(&borrows).Error; e != nil {
		return nil, e
	}

	return borrows, nil
}

func (s *borrowStore) CountOfBorrowers(ctx context.Context, symbol string) (int64, error) {
	var count int64
	if e := s.db.View().Model(core.Borrow{}).Select("count(user_id)").Where("symbol=?", symbol).Row().Scan(&count); e != nil {
		return 0, e
	}

	return count, nil
}

func (s *borrowStore) Users(ctx context.Context) ([]string, error) {
	var users []string
	if e := s.db.View().Model(core.Borrow{}).Select("distinct user_id").Pluck("user_id", &users).Error; e != nil {
		return nil, e
	}

	return users, nil
}
