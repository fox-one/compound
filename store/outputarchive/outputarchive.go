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

func (s *store) Save(ctx context.Context, archive *core.OutputArchive) error {
	return s.db.Update().Where("trace_id=?", archive.TraceID).FirstOrCreate(archive).Error
}

func (s *store) Find(ctx context.Context, traceID string) (*core.OutputArchive, error) {
	var archive core.OutputArchive
	if err := s.db.View().Where("trace_id=?", traceID).First(&archive).Error; err != nil {
		return nil, err
	}

	return &archive, nil
}
