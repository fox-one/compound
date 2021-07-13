package borrow

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
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

func (s *borrowStore) Create(ctx context.Context, borrow *core.Borrow) error {
	if e := s.db.Update().Where("user_id=? and asset_id=?", borrow.UserID, borrow.AssetID).Create(borrow).Error; e != nil {
		return e
	}

	return nil
}

func (s *borrowStore) Find(ctx context.Context, userID string, assetID string) (*core.Borrow, error) {
	var borrow core.Borrow
	if e := s.db.View().Where("user_id=? and asset_id=?", userID, assetID).First(&borrow).Error; e != nil {
		if gorm.IsRecordNotFoundError(e) {
			return &core.Borrow{}, nil
		}
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

func (s *borrowStore) FindByAssetID(ctx context.Context, assetID string) ([]*core.Borrow, error) {
	var borrows []*core.Borrow
	if e := s.db.View().Where("asset_id=?", assetID).Find(&borrows).Error; e != nil {
		return nil, e
	}

	return borrows, nil
}

func (s *borrowStore) Update(ctx context.Context, borrow *core.Borrow, version int64) error {
	if version > borrow.Version {
		oldVersion := borrow.Version
		borrow.Version = version
		return s.db.Update().Model(borrow).Where("version=?", oldVersion).Updates(borrow).Error
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

func (s *borrowStore) CountOfBorrowers(ctx context.Context, assetID string) (int64, error) {
	var count int64
	if e := s.db.View().Model(core.Borrow{}).Select("count(user_id)").Where("asset_id=?", assetID).Row().Scan(&count); e != nil {
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
