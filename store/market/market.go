package market

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type marketStore struct {
	db *db.DB
}

// New new market store
func New(db *db.DB) core.IMarketStore {
	return &marketStore{db: db}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Market{})
		if err := tx.AutoMigrate(core.Market{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *marketStore) Create(ctx context.Context, market *core.Market) error {
	return s.db.Update().Where("asset_id=?", market.AssetID).FirstOrCreate(market).Error
}

func (s *marketStore) Find(ctx context.Context, assetID string) (*core.Market, error) {
	if assetID == "" {
		return &core.Market{}, nil
	}

	var market core.Market
	if err := s.db.View().Where("asset_id=?", assetID).First(&market).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &core.Market{}, nil
		}
		return nil, err
	}

	return afterFind(&market), nil
}

func (s *marketStore) FindBySymbol(ctx context.Context, symbol string) (*core.Market, error) {
	if symbol == "" {
		return &core.Market{}, nil
	}

	var market core.Market
	if err := s.db.View().Where("symbol=?", symbol).First(&market).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &core.Market{}, nil
		}
		return nil, err
	}

	return afterFind(&market), nil
}

func (s *marketStore) FindByCToken(ctx context.Context, ctokenAssetID string) (*core.Market, error) {
	if ctokenAssetID == "" {
		return &core.Market{}, nil
	}

	var market core.Market
	if err := s.db.View().Where("c_token_asset_id=?", ctokenAssetID).First(&market).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &core.Market{}, nil
		}
		return nil, err
	}

	return afterFind(&market), nil
}

func (s *marketStore) All(ctx context.Context) ([]*core.Market, error) {
	var markets []*core.Market
	if err := s.db.View().Find(&markets).Error; err != nil {
		return nil, err
	}
	for _, m := range markets {
		afterFind(m)
	}
	return markets, nil
}

func (s *marketStore) AllAsMap(ctx context.Context) (map[string]*core.Market, error) {
	markets, e := s.All(ctx)
	if e != nil {
		return nil, e
	}

	maps := make(map[string]*core.Market)

	for _, m := range markets {
		afterFind(m)
	}

	return maps, nil
}

func (s *marketStore) Update(ctx context.Context, market *core.Market, version int64) error {
	if version > market.Version {
		// do real update
		oldVersion := market.Version
		market.Version = version
		tx := s.db.Update().Model(market).Where("version=?", oldVersion).Updates(market)

		if tx.Error != nil {
			return tx.Error
		}

		if tx.RowsAffected == 0 {
			return db.ErrOptimisticLock
		}
	}

	return nil
}

func afterFind(market *core.Market) *core.Market {
	if !market.ExchangeRate.IsPositive() {
		market.ExchangeRate = decimal.New(1, 0)
	}
	if !market.BorrowIndex.IsPositive() {
		market.BorrowIndex = decimal.New(1, 0)
	}
	return market
}
