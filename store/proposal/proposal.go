package proposal

import (
	"context"

	"compound/core"

	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
)

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(core.Proposal{})

		if err := tx.AutoMigrate(core.Proposal{}).Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_proposal_trace", "trace_id").Error; err != nil {
			return err
		}

		return nil
	})
}

// New new proposal store
func New(db *db.DB) core.ProposalStore {
	return &proposalStore{db: db}
}

type proposalStore struct {
	db *db.DB
}

func (s *proposalStore) Create(ctx context.Context, proposal *core.Proposal) error {
	return s.db.Update().Where("trace_id = ?", proposal.TraceID).FirstOrCreate(proposal).Error
}

func (s *proposalStore) Find(ctx context.Context, trace string) (*core.Proposal, bool, error) {
	var proposal core.Proposal
	if err := s.db.View().Where("trace_id = ?", trace).First(&proposal).Error; err != nil {
		return nil, gorm.IsRecordNotFoundError(err), err
	}

	return &proposal, false, nil
}

func toUpdateParams(proposal *core.Proposal) map[string]interface{} {
	return map[string]interface{}{
		"passed_at": proposal.PassedAt,
		"votes":     proposal.Votes,
	}
}

func (s *proposalStore) Update(ctx context.Context, proposal *core.Proposal, version int64) error {
	updates := toUpdateParams(proposal)
	updates["version"] = version

	tx := s.db.Update().Model(proposal).Where("version = ?", proposal.Version).Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return db.ErrOptimisticLock
	}

	return nil
}

func (s *proposalStore) List(ctx context.Context, fromID int64, limit int) ([]*core.Proposal, error) {
	var proposals []*core.Proposal
	if err := s.db.View().Where("id > ?", fromID).Limit(limit).Find(&proposals).Error; err != nil {
		return nil, err
	}

	return proposals, nil
}
