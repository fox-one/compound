package supply

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
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

func (s *supplyStore) Save(ctx context.Context, tx *db.DB, supply *core.Supply) error {
	if e := tx.Update().Where("user_id=? and ctoken_asset_id=?", supply.UserID, supply.CTokenAssetID).FirstOrCreate(supply).Error; e != nil {
		return e
	}

	return nil
}
func (s *supplyStore) Find(ctx context.Context, userID string, symbol string) (*core.Supply, error) {
	var supply core.Supply
	if e := s.db.View().Where("user_id=? and symbol=?", userID, symbol).First(&supply).Error; e != nil {
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

func (s *supplyStore) Update(ctx context.Context, tx *db.DB, supply *core.Supply) error {
	version := supply.Version
	supply.Version++
	if err := tx.Update().Model(core.Supply{}).Where("user_id=? ctoken_asset_id=? and version=?", supply.UserID, supply.CTokenAssetID, version).Updates(supply).Error; err != nil {
		return err
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
	if e := s.db.View().Where("ctoken_asset_id=?", assetID).Find(&supplies).Error; e != nil {
		return nil, e
	}

	return supplies, nil
}
func (s *supplyStore) SumOfSupplies(ctx context.Context, symbol string) (decimal.Decimal, error) {
	var sum decimal.Decimal
	if e := s.db.View().Model(core.Supply{}).Select("sum(principal)").Where("symbol=?", symbol).Row().Scan(&sum); e != nil {
		return decimal.Zero, e
	}

	return sum, nil

}

func (s *supplyStore) CountOfSupplies(ctx context.Context, symbol string) (int64, error) {
	var count int64
	if e := s.db.View().Model(core.Supply{}).Select("count(user_id)").Where("symbol=?", symbol).Row().Scan(&count); e != nil {
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
