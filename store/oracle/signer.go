package oracle

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

type oracleSignerStore struct {
	db *db.DB
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.OracleSigner{})
		if err := tx.AutoMigrate(core.OracleSigner{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func NewSignerStore(db *db.DB) core.OracleSignerStore {
	return &oracleSignerStore{db: db}
}

func (s *oracleSignerStore) Save(ctx context.Context, userID, publicKey string) error {
	feed := core.OracleSigner{
		UserID:    userID,
		PublicKey: publicKey,
	}

	return s.db.Update().Where("user_id = ?", userID).Assign(core.OracleSigner{UserID: userID, PublicKey: publicKey}).FirstOrCreate(&feed).Error
}

func (s *oracleSignerStore) Delete(ctx context.Context, userID string) error {
	return s.db.Update().Where("user_id = ?", userID).Delete(core.OracleSigner{}).Error
}

func (s *oracleSignerStore) FindAll(ctx context.Context) ([]*core.OracleSigner, error) {
	var signers []*core.OracleSigner
	if err := s.db.View().Find(&signers).Error; err != nil {
		return nil, err
	}

	return signers, nil
}
