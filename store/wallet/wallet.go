package wallet

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"sync/atomic"

	"compound/core"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
)

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Output{})
		if err := tx.AutoMigrate(core.Output{}).Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_outputs_trace", "trace_id").Error; err != nil {
			return err
		}

		if err := tx.AddIndex("idx_outputs_asset_transfer", "asset_id", "spent_by").Error; err != nil {
			return err
		}

		return nil
	})

	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Transfer{})
		if err := tx.AutoMigrate(core.Transfer{}).Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_transfers_trace", "trace_id").Error; err != nil {
			return err
		}

		if err := tx.AddIndex("idx_transfers_handled_passed", "handled", "passed").Error; err != nil {
			return err
		}

		return nil
	})

	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.RawTransaction{})
		if err := tx.AutoMigrate(core.RawTransaction{}).Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_raw_transactions_trace", "trace_id").Error; err != nil {
			return err
		}

		return nil
	})
}

func New(db *db.DB) core.WalletStore {
	return &walletStore{db: db}
}

type walletStore struct {
	db          *db.DB
	once        sync.Once
	rawOutputID int64
}

func afterFindOutput(output *core.Output) *core.Output {
	var utxo mixin.MultisigUTXO
	if err := json.Unmarshal(output.Data, &utxo); err == nil {
		output.UTXO = &utxo
	}

	return output
}

func save(db *db.DB, output *core.Output, ack bool) error {
	tx := db.Update().Model(output).Where("trace_id = ?", output.TraceID).Updates(map[string]interface{}{
		"data":    output.Data,
		"state":   output.State,
		"version": gorm.Expr("version + 1"),
	})
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		if ack {
			return db.Update().Create(output).Error
		}

		return saveRawOutput(db, output)
	}

	return nil
}

func (s *walletStore) Save(ctx context.Context, outputs []*core.Output, end bool) error {
	s.once.Do(func() {
		go func() {
			err := s.runSync(ctx)
			logger.FromContext(ctx).WithError(err).Infoln("runSync end")
		}()
	})

	if err := s.db.Tx(func(tx *db.DB) error {
		for _, utxo := range outputs {
			if err := save(tx, utxo, false); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if end {
		id, err := findLastRawOutputID(s.db)
		if err != nil {
			return err
		}

		if id > 0 {
			atomic.StoreInt64(&s.rawOutputID, id)
		}
	}

	return nil
}

func (s *walletStore) List(_ context.Context, fromID int64, limit int) ([]*core.Output, error) {
	var outputs []*core.Output
	if err := s.db.View().
		Where("id > ?", fromID).
		Limit(limit).
		Order("id").
		Find(&outputs).Error; err != nil {
		return nil, err
	}

	for _, output := range outputs {
		afterFindOutput(output)
	}

	return outputs, nil
}

func (s *walletStore) FindSpentBy(ctx context.Context, assetID, spentBy string) (*core.Output, error) {
	var output core.Output
	if err := s.db.View().Where("asset_id = ? AND spent_by = ?", assetID, spentBy).Take(&output).Error; err != nil {
		return nil, err
	}

	return afterFindOutput(&output), nil
}

func (s *walletStore) ListSpentBy(ctx context.Context, assetID string, spentBy string) ([]*core.Output, error) {
	var outputs []*core.Output
	if err := s.db.View().
		Where("asset_id = ? AND spent_by = ?", assetID, spentBy).
		Order("id").
		Find(&outputs).Error; err != nil {
		return nil, err
	}

	for _, output := range outputs {
		afterFindOutput(output)
	}

	return outputs, nil
}

func (s *walletStore) ListUnspent(_ context.Context, assetID string, limit int) ([]*core.Output, error) {
	var outputs []*core.Output
	if err := s.db.View().
		Where("asset_id = ? AND spent_by = ?", assetID, "").
		Limit(limit).
		Order("id").
		Find(&outputs).Error; err != nil {
		return nil, err
	}

	for _, output := range outputs {
		afterFindOutput(output)
	}

	return outputs, nil
}

func afterFindTransfer(transfer *core.Transfer) {
	if transfer.Threshold == 0 {
		transfer.Threshold = uint8(len(transfer.Opponents))
	}
}

func (s *walletStore) CreateTransfers(_ context.Context, transfers []*core.Transfer) error {
	sort.Slice(transfers, func(i, j int) bool {
		return transfers[i].TraceID < transfers[j].TraceID
	})

	return s.db.Tx(func(tx *db.DB) error {
		for _, transfer := range transfers {
			if err := tx.Update().Where("trace_id = ?", transfer.TraceID).FirstOrCreate(transfer).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func updateTransfer(db *db.DB, transfer *core.Transfer) error {
	return db.Update().Model(transfer).Updates(map[string]interface{}{
		"assigned": transfer.Assigned,
		"handled":  transfer.Handled,
		"passed":   transfer.Passed,
	}).Error
}

func (s *walletStore) UpdateTransfer(ctx context.Context, transfer *core.Transfer) error {
	return updateTransfer(s.db, transfer)
}

func (s *walletStore) ListTransfers(ctx context.Context, status core.TransferStatus, limit int) ([]*core.Transfer, error) {
	query := s.db.View().Limit(limit).Order("id")

	switch status {
	case core.TransferStatusPending:
		query = query.Where("handled = ? AND assigned = ?", 0, 0)
	case core.TransferStatusAssigned:
		query = query.Where("handled = ? AND assigned = ?", 0, 1)
	case core.TransferStatusHandled:
		query = query.Where("handled = ? AND passed = ?", 1, 0)
	default:
		query = query.Where("handled = ? AND passed = ?", 1, 1)
	}

	var transfers []*core.Transfer
	if err := query.Find(&transfers).Error; err != nil {
		return nil, err
	}

	for _, t := range transfers {
		afterFindTransfer(t)
	}

	return transfers, nil
}

func (s *walletStore) Assign(_ context.Context, outputs []*core.Output, transfer *core.Transfer) error {
	ids := make([]int64, 0, len(outputs))
	for _, output := range outputs {
		ids = append(ids, output.ID)
	}

	return s.db.Tx(func(tx *db.DB) error {
		if err := tx.Update().Model(core.Output{}).
			Where("id IN (?)", ids).
			Update("spent_by", transfer.TraceID).Error; err != nil {
			return err
		}

		transfer.Assigned = true
		if transfer.ID > 0 {
			if err := updateTransfer(tx, transfer); err != nil {
				return err
			}
		} else {
			if err := tx.Update().Create(transfer).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *walletStore) CreateRawTransaction(_ context.Context, tx *core.RawTransaction) error {
	return s.db.Update().Where("trace_id = ?", tx.TraceID).FirstOrCreate(tx).Error
}

func (s *walletStore) ListPendingRawTransactions(_ context.Context, limit int) ([]*core.RawTransaction, error) {
	var txs []*core.RawTransaction
	if err := s.db.View().Limit(limit).Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (s *walletStore) ExpireRawTransaction(_ context.Context, tx *core.RawTransaction) error {
	return s.db.Update().Model(tx).Where("id = ?", tx.ID).Delete(tx).Error
}

func (s *walletStore) CountOutputs(ctx context.Context) (int64, error) {
	var output core.Output
	if err := s.db.View().Select("id").Last(&output).Error; err != nil && !db.IsErrorNotFound(err) {
		return 0, err
	}

	return output.ID, nil
}

func (s *walletStore) CountUnhandledTransfers(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.View().Model(core.Transfer{}).Where("handled = ?", 0).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
