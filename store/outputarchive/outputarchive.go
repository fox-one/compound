package outputarchive

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

type store struct {
	db *db.DB
}

// New new outputarchive store
func New(db *db.DB) core.OutputArchiveStore {
	return &store{
		db: db,
	}
}

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.OutputArchive{})
		if err := tx.AutoMigrate(core.OutputArchive{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *store) Save(ctx context.Context, tx *db.DB, archive *core.OutputArchive) error {
	return tx.Update().Where("trace_id=?", archive.TraceID).FirstOrCreate(archive).Error
}

func (s *store) Find(ctx context.Context, traceID string) (*core.OutputArchive, error) {
	var archive core.OutputArchive
	if err := s.db.View().Where("trace_id=?", traceID).First(&archive).Error; err != nil {
		return nil, err
	}

	return &archive, nil
}
