package transfer

// type transferStore struct {
// 	db *db.DB
// }

// // New new transfer store
// func New(db *db.DB) core.ITransferStore {
// 	return &transferStore{
// 		db: db,
// 	}
// }

// func init() {
// 	db.RegisterMigrate(func(db *db.DB) error {
// 		tx := db.Update().Model(core.Transfer{})
// 		if err := tx.AutoMigrate(core.Transfer{}).Error; err != nil {
// 			return err
// 		}

// 		return nil
// 	})
// }

// func (s *transferStore) Create(ctx context.Context, tx *db.DB, transfer *core.Transfer) error {
// 	transfer.Status = core.TransferStatusPending
// 	return tx.Update().Where("trace_id=?", transfer.TraceID).FirstOrCreate(transfer).Error
// }
// func (s *transferStore) Delete(ctx context.Context, tx *db.DB, ids ...uint64) error {
// 	return tx.Update().Where("id in (?)", ids).Delete(core.Transfer{}).Error
// }
// func (s *transferStore) Top(ctx context.Context, limit int) ([]*core.Transfer, error) {
// 	if limit <= 0 {
// 		return nil, errors.New("invalid limit")
// 	}

// 	var transfers []*core.Transfer
// 	if e := s.db.View().Limit(limit).Offset(0).Order("created_at ASC").Find(&transfers).Error; e != nil {
// 		return nil, e
// 	}

// 	return transfers, nil
// }

// func (s *transferStore) FindByStatus(ctx context.Context, status string) ([]*core.Transfer, error) {
// 	var transfers []*core.Transfer
// 	if e := s.db.View().Where("status=?", status).Find(&transfers).Error; e != nil {
// 		return nil, e
// 	}
// 	return transfers, nil
// }

// func (s *transferStore) UpdateStatus(ctx context.Context, tx *db.DB, id uint64, status string) error {
// 	return s.db.Update().Model(core.Transfer{}).Where("id=?", id).Update("status", status).Error
// }

// func (s *transferStore) DeleteByTime(t time.Time) error {
// 	return s.db.Update().Where("status=? and created_at < ?", core.TransferStatusDone, t).Delete(core.Transfer{}).Error
// }
