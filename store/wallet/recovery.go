package wallet

import (
	"compound/core"
	"compound/internal/mixinet"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx/types"
)

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(RawOutput{})

		if err := tx.AutoMigrate(RawOutput{}).Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_raw_outputs_trace", "trace_id").Error; err != nil {
			return err
		}

		if err := tx.AddIndex("idx_raw_outputs_created", "created_at").Error; err != nil {
			return err
		}

		return nil
	})
}

type RawOutput struct {
	ID        int64          `sql:"PRIMARY_KEY" json:"id,omitempty"`
	CreatedAt int64          `json:"created_at"`
	TraceID   string         `sql:"size:36" json:"trace_id"`
	Version   int64          `sql:"not null" json:"version"`
	Data      types.JSONText `sql:"type:TEXT" json:"data"`
}

func saveRawOutput(db *db.DB, output *core.Output) error {
	data, _ := json.Marshal(output)

	raw := &RawOutput{
		CreatedAt: output.CreatedAt.UnixNano(),
		TraceID:   output.TraceID,
		Version:   1,
		Data:      data,
	}

	tx := db.Update().Model(raw).
		Where("trace_id = ?", raw.TraceID).
		Updates(map[string]interface{}{
			"data":    raw.Data,
			"version": gorm.Expr("version + 1"),
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return db.Update().Create(raw).Error
	}

	return nil
}

func findLastRawOutputID(db *db.DB) (int64, error) {
	var outputs []*RawOutput
	if err := db.View().Model(RawOutput{}).Select("id").Order("id DESC").Limit(1).Find(&outputs).Error; err != nil {
		return 0, err
	}

	if len(outputs) == 0 {
		return 0, nil
	}

	return outputs[0].ID, nil
}

func (s *walletStore) runSync(ctx context.Context) error {
	dur := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			const limit = 500

			id := atomic.LoadInt64(&s.rawOutputID)
			n, err := syncRawOutputs(s.db, id, limit)
			if err != nil {
				logger.FromContext(ctx).WithError(err).Errorln("syncRawOutputs")
				dur = time.Second
			} else if n == 0 {
				dur = 600 * time.Millisecond
			} else {
				dur = 300 * time.Millisecond
			}
		}
	}
}

func syncRawOutputs(tx *db.DB, id int64, limit int) (int, error) {
	var raws []*RawOutput
	if err := tx.View().Where("id <= ?", id).Order("created_at").Limit(limit).Find(&raws).Error; err != nil {
		return 0, err
	}

	if len(raws) == 0 {
		return 0, nil
	}

	if len(raws) == limit {
		raws = trimSuffix(raws)
	}

	outputs := make([]*core.Output, 0, len(raws))
	ids := make([]int64, 0, len(raws))

	for _, raw := range raws {
		var output core.Output
		if err := json.Unmarshal(raw.Data, &output); err != nil {
			return 0, fmt.Errorf("unmarshal RawOutput failed: %w", err)
		}

		outputs = append(outputs, &output)
		ids = append(ids, raw.ID)
	}

	mixinet.SortOutputs(outputs)

	if err := tx.Tx(func(tx *db.DB) error {
		for _, output := range outputs {
			if err := save(tx, output, true); err != nil {
				return err
			}
		}

		return tx.Update().Where("id IN (?)", ids).Delete(RawOutput{}).Error
	}); err != nil {
		return 0, err
	}

	return len(outputs), nil
}

func trimSuffix(raws []*RawOutput) []*RawOutput {
	var (
		r = len(raws) - 1
		l = r - 1
	)

	for l >= 0 {
		if raws[l].CreatedAt != raws[r].CreatedAt {
			break
		}

		l = l - 1
	}

	if l >= 0 {
		raws = raws[:l+1]
	}

	return raws
}
