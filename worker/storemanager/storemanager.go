package storemanager

import (
	"compound/core"
	"compound/worker"
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

// Worker store manager worker
type Worker struct {
	worker.BaseJob
	Config        *core.Config
	TransferStore core.ITransferStore
	SnapshotStore core.ISnapshotStore
}

// New new block worker
func New(config *core.Config, transferStr core.ITransferStore, snapshotStr core.ISnapshotStore) *Worker {
	job := Worker{
		Config:        config,
		TransferStore: transferStr,
		SnapshotStore: snapshotStr,
	}

	l, _ := time.LoadLocation(job.Config.App.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 600s"
	job.Cron.AddFunc(spec, job.Run)
	job.OnWork = func() error {
		return job.onWork(context.Background())
	}

	return &job
}

func (w *Worker) onWork(ctx context.Context) error {
	now := time.Now()
	checkPoint := now.AddDate(0, 0, -2)

	w.TransferStore.DeleteByTime(checkPoint)
	w.SnapshotStore.DeleteByTime(checkPoint)
	return nil
}
