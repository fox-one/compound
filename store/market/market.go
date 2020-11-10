package market

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/store/db"
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

func (s *marketStore) Save(ctx context.Context, tx *db.DB, market *core.Market) error {
	if err := tx.Update().Create(market).Error; err != nil {
		return err
	}
	return nil
}
func (s *marketStore) Find(ctx context.Context, assetID, symbol string) (*core.Market, error) {
	if assetID == "" && symbol == "" {
		return nil, errors.New("invalid asset_id and symbol")
	}

	var market core.Market
	query := s.db.View()
	if assetID != "" {
		query = query.Where("asset_id=?", assetID)
	}
	if symbol != "" {
		query = query.Where("symbol=?", symbol)
	}

	if err := query.First(&market).Error; err != nil {
		return nil, err
	}

	return &market, nil
}

func (s *marketStore) FindByCToken(ctx context.Context, ctokenAssetID, ctokenSymbol string) (*core.Market, error) {
	if ctokenAssetID == "" && ctokenSymbol == "" {
		return nil, errors.New("invalid asset_id and symbol")
	}

	var market core.Market
	query := s.db.View()
	if ctokenAssetID != "" {
		query = query.Where("asset_id=?", ctokenAssetID)
	}
	if ctokenSymbol != "" {
		query = query.Where("symbol=?", ctokenSymbol)
	}

	if err := query.First(&market).Error; err != nil {
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

func (s *marketStore) Update(ctx context.Context, tx *db.DB, market *core.Market) error {
	version := market.Version
	market.Version++
	if err := tx.Update().Model(core.Market{}).Where("asset_id=? and version=?", market.AssetID, version).Update(market).Error; err != nil {
		return err
	}

	return nil
}
