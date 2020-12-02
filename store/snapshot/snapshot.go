package snapshot

import (
	"compound/core"
	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
)

type snapshotStore struct {
	db *db.DB
}

// New new snapshot store instance
func New(db *db.DB) core.ISnapshotStore {
	return &snapshotStore{
		db: db,
	}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Snapshot{})
		if err := tx.AutoMigrate(core.Snapshot{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *snapshotStore) Save(ctx context.Context, snapshot *core.Snapshot) error {
	return s.db.Update().Where("snapshot_id=?", snapshot.SnapshotID).FirstOrCreate(snapshot).Error
}

func (s *snapshotStore) Find(ctx context.Context, snapshotID string) (*core.Snapshot, error) {
	var snapshot core.Snapshot
	if e := s.db.View().Where("snapshot_id=?", snapshotID).First(&snapshot).Error; e != nil {
		return nil, e
	}

	return &snapshot, nil
}

func (s *snapshotStore) DeleteByTime(t time.Time) error {
	return s.db.Update().Where("created_at < ?", t).Delete(core.Snapshot{}).Error
}
