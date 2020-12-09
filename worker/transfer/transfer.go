package transfer

import (
	"compound/core"
	"compound/worker"

	"context"
	"time"

	"github.com/fox-one/pkg/store/db"
	"github.com/robfig/cron/v3"
)

// Worker worker
type Worker struct {
	worker.BaseJob
	DB         *db.DB
	MainWallet *core.Wallet
	Config     *core.Config
}

// New new block worker
func New(db *db.DB, mainWallet *core.Wallet, cfg *core.Config) *Worker {
	job := Worker{
		DB:         db,
		MainWallet: mainWallet,
		Config:     cfg,
	}

	l, _ := time.LoadLocation(job.Config.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 1s"
	job.Cron.AddFunc(spec, job.Run)
	job.OnWork = func() error {
		return job.onWork(context.Background())
	}

	return &job
}

func (w *Worker) onWork(ctx context.Context) error {
	// TODO should be deleted
	// pendingTransfers, e := w.TransferStore.FindByStatus(ctx, core.TransferStatusPending)
	// if e != nil {
	// 	return e
	// }

	// for _, transfer := range pendingTransfers {
	// 	w.doTransfer(ctx, transfer)
	// }

	return nil
}

func (w *Worker) doTransfer(ctx context.Context, transfer *core.Transfer) error {
	//TODO should be deleted
	// log := logger.FromContext(ctx).WithField("worker", "transfer")
	// return w.DB.Tx(func(tx *db.DB) error {
	// 	input := mixin.TransferInput{
	// 		AssetID:    transfer.AssetID,
	// 		OpponentID: transfer.OpponentID,
	// 		Amount:     transfer.Amount,
	// 		TraceID:    transfer.TraceID,
	// 	}

	// 	input.Memo = transfer.Memo
	// 	if _, e := w.MainWallet.Client.Transfer(ctx, &input, w.MainWallet.Pin); e != nil {
	// 		log.Errorln(e)
	// 		return e
	// 	}
	// 	//update transfer status
	// 	if e := w.TransferStore.UpdateStatus(ctx, tx, transfer.ID, core.TransferStatusDone); e != nil {
	// 		log.Errorln(e)
	// 		return e
	// 	}

	// 	return nil
	// })
	return nil
}
