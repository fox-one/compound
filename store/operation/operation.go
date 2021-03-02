package operation

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

type allowListStore struct {
	db *db.DB
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.AllowList{})
		if err := tx.AutoMigrate(core.AllowList{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// NewAllowListStore new allowlist store
func NewAllowListStore(db *db.DB) core.IAllowListStore {
	return &allowListStore{
		db: db,
	}
}

func (s *allowListStore) Create(ctx context.Context, allowList *core.AllowList) error {
	return s.db.Update().Where("user_id=? and scope=?", allowList.UserID, allowList.Scope).FirstOrCreate(allowList).Error
}

func (s *allowListStore) Find(ctx context.Context, userID string, scope core.OperationScope) (*core.AllowList, error) {
	var allowList core.AllowList
	if e := s.db.View().Where("user_id=? and scope=?", userID, scope).First(&allowList).Error; e != nil {
		return nil, e
	}
	return &allowList, nil
}

func (s *allowListStore) Delete(ctx context.Context, userID string, scope core.OperationScope) error {
	return s.db.Update().Where("user_id=? and scope=?", userID, scope).Delete(core.AllowList{}).Error
}
