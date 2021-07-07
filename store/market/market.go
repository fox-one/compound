package market

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
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

func (s *marketStore) Save(ctx context.Context, market *core.Market) error {
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

	return &market, nil
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

	return &market, nil
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

	return &market, nil
}

func (s *marketStore) All(ctx context.Context) ([]*core.Market, error) {
	var markets []*core.Market
	if err := s.db.View().Find(&markets).Error; err != nil {
		return nil, err
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
		maps[m.Symbol] = m
	}

	return maps, nil
}

func (s *marketStore) Update(ctx context.Context, market *core.Market, version int64) error {
	if version > market.Version {
		// do real update
		market.Version = version
		return s.db.Update().Model(market).Updates(market).Error
	}

	return nil
}
