package snapshot

import (
	"compound/core"
	"context"
	"errors"
	"time"

	"github.com/bluele/gcache"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
)

// SnapshotWorker snapshot worker
type SnapshotWorker struct {
	config        *core.Config
	property      property.Store
	walletService core.IWalletService
	snapshotCache gcache.Cache
}

const (
	checkPointKey = "compound_snapshot_checkpoint"
	limit         = 500
)

// New new snapshot worker
func New(
	config *core.Config,
	property property.Store,
	walletService core.IWalletService,
) *SnapshotWorker {
	return &SnapshotWorker{
		config:        config,
		property:      property,
		walletService: walletService,
		snapshotCache: gcache.New(limit).LRU().Build(),
	}
}

// Run run snapshot worker
func (w *SnapshotWorker) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "snapshot")
	ctx = logger.WithContext(ctx, log)

	duration := time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(duration):
			if err := w.realRun(ctx); err == nil {
				duration = 100 * time.Millisecond
			} else {
				duration = time.Second
			}
		}
	}
}

func (w *SnapshotWorker) realRun(ctx context.Context) error {
	log := logger.FromContext(ctx)
	checkPoint, err := w.property.Get(ctx, checkPointKey)
	if err != nil {
		log.WithError(err).Errorf("read property error: %s", checkPointKey)
		return err
	}

	snapshots, next, err := w.walletService.PullSnapshots(ctx, checkPoint.String(), limit)
	if err != nil {
		log.WithError(err).Error("pull snapshots error")
		return err
	}

	if len(snapshots) == 0 {
		return errors.New("no more snapshots")
	}

	for _, snapshot := range snapshots {
		if snapshot.UserID == "" {
			continue
		}

		if snapshot.UserID != w.config.Mixin.ClientID {
			continue
		}

		if snapshot.Amount.IsNegative() {
			continue
		}

		if w.snapshotCache.Has(snapshot.ID) {
			continue
		}

		if err := w.handleSnapshot(ctx, snapshot); err != nil {
			return err
		}

		w.snapshotCache.Set(snapshot.ID, nil)
	}

	if checkPoint.String() != next {
		if err := w.property.Save(ctx, checkPointKey, next); err != nil {
			log.WithError(err).Errorf("update property error: %s", checkPointKey)
			return err
		}
	}

	return nil
}

func (w *SnapshotWorker) handleSnapshot(ctx context.Context, snapshot *core.Snapshot) error {
	//TODO add handle snapshot code here
	return nil
}
