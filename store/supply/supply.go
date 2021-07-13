package supply

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type supplyStore struct {
	db *db.DB
}

// New new supply store
func New(db *db.DB) core.ISupplyStore {
	return &supplyStore{
		db: db,
	}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Supply{})
		if err := tx.AutoMigrate(core.Supply{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *supplyStore) Create(ctx context.Context, supply *core.Supply) error {
	if e := s.db.Update().Where("user_id=? and c_token_asset_id=?", supply.UserID, supply.CTokenAssetID).Create(supply).Error; e != nil {
		return e
	}

	return nil
}

func (s *supplyStore) Find(ctx context.Context, userID string, ctokenAssetID string) (*core.Supply, error) {
	var supply core.Supply
	if e := s.db.View().Where("user_id=? and c_token_asset_id=?", userID, ctokenAssetID).First(&supply).Error; e != nil {
		if gorm.IsRecordNotFoundError(e) {
			return &core.Supply{}, nil
		}

		return nil, e
	}

	return &supply, nil
}

func (s *supplyStore) FindByUser(ctx context.Context, userID string) ([]*core.Supply, error) {
	var supplies []*core.Supply
	if e := s.db.View().Where("user_id=?", userID).Find(&supplies).Error; e != nil {
		return nil, e
	}

	return supplies, nil
}

func (s *supplyStore) Update(ctx context.Context, supply *core.Supply, version int64) error {
	if version > supply.Version {
		oldVersion := supply.Version
		supply.Version = version
		tx := s.db.Update().Model(supply).Where("version=?", oldVersion).Updates(supply)

		if tx.Error != nil {
			return tx.Error
		}

		if tx.RowsAffected == 0 {
			return db.ErrOptimisticLock
		}
	}

	return nil
}

func (s *supplyStore) All(ctx context.Context) ([]*core.Supply, error) {
	var supplies []*core.Supply
	if e := s.db.View().Find(&supplies).Error; e != nil {
		return nil, e
	}

	return supplies, nil
}

func (s *supplyStore) FindByCTokenAssetID(ctx context.Context, assetID string) ([]*core.Supply, error) {
	var supplies []*core.Supply
	if e := s.db.View().Where("c_token_asset_id=?", assetID).Find(&supplies).Error; e != nil {
		return nil, e
	}

	return supplies, nil
}
func (s *supplyStore) SumOfSupplies(ctx context.Context, ctokenAssetID string) (decimal.Decimal, error) {
	var sum decimal.Decimal
	if e := s.db.View().Model(core.Supply{}).Select("sum(collaterals)").Where("c_token_asset_id=?", ctokenAssetID).Row().Scan(&sum); e != nil {
		return decimal.Zero, e
	}

	return sum, nil

}

func (s *supplyStore) CountOfSuppliers(ctx context.Context, ctokenAssetID string) (int64, error) {
	var count int64
	if e := s.db.View().Model(core.Supply{}).Select("count(user_id)").Where("c_token_asset_id=?", ctokenAssetID).Row().Scan(&count); e != nil {
		return 0, e
	}

	return count, nil
}

func (s *supplyStore) Users(ctx context.Context) ([]string, error) {
	var users []string
	if e := s.db.View().Model(core.Supply{}).Select("distinct user_id").Pluck("user_id", &users).Error; e != nil {
		return nil, e
	}

	return users, nil
}
