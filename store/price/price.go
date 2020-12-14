package price

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

type priceStore struct {
	db *db.DB
}

// New new price store
func New(db *db.DB) core.IPriceStore {
	return &priceStore{
		db: db,
	}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Price{})

		if err := tx.AutoMigrate(core.Price{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *priceStore) Create(ctx context.Context, tx *db.DB, price *core.Price) error {
	return tx.Update().Where("asset_id=? and block_number=?", price.AssetID, price.BlockNumber).FirstOrCreate(price).Error
}

func (s *priceStore) FindByAssetBlock(ctx context.Context, assetID string, blockNumber int64) (*core.Price, error) {
	var price core.Price
	if e := s.db.View().Where("asset_id=? and block_number=?", assetID, blockNumber).Find(&price).Error; e != nil {
		return nil, e
	}
	return &price, nil
}

func (s *priceStore) Update(ctx context.Context, tx *db.DB, price *core.Price) error {
	version := price.Version
	price.Version++
	return tx.Update().Model(core.Price{}).Where("asset_id=? and block_number=? and version=?", price.AssetID, price.BlockNumber, version).Updates(price).Error
}
