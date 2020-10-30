package supply

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
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

func (s *supplyStore) Save(ctx context.Context, supply *core.Supply) error {
	if e := s.db.Update().Where("user_id=? and symbol=?", supply.UserID, supply.Symbol).FirstOrCreate(supply).Error; e != nil {
		return e
	}

	return nil
}
func (s *supplyStore) Find(ctx context.Context, userID string, symbols ...string) ([]*core.Supply, error) {
	query := s.db.View().Where("user_id=?", userID)
	if len(symbols) > 0 {
		query = query.Where("symbol in (?)", symbols)
	}
	var supplies []*core.Supply
	if e := query.Find(&supplies).Error; e != nil {
		return nil, e
	}

	return supplies, nil
}

func (s *supplyStore) Update(ctx context.Context, tx *db.DB, supply *core.Supply) error {
	version := supply.Version
	supply.Version++
	if err := tx.Update().Model(core.Supply{}).Where("user_id=? symbol=? and version=?", supply.UserID, supply.Symbol, version).Updates(supply).Error; err != nil {
		return err
	}

	return nil
}
