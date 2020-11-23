package transfer

import (
	"compound/core"
	"compound/worker"

	"context"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
	"github.com/robfig/cron/v3"
)

// Worker worker
type Worker struct {
	worker.BaseJob
	DB            *db.DB
	MainWallet    *core.Wallet
	Config        *core.Config
	TransferStore core.ITransferStore
	WalletService core.IWalletService
}

// New new block worker
func New(db *db.DB, mainWallet *core.Wallet, cfg *core.Config, transferStr core.ITransferStore, walletSrv core.IWalletService) *Worker {
	job := Worker{
		DB:            db,
		MainWallet:    mainWallet,
		Config:        cfg,
		TransferStore: transferStr,
		WalletService: walletSrv,
	}

	l, _ := time.LoadLocation(job.Config.App.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 1s"
	job.Cron.AddFunc(spec, job.Run)
	job.OnWork = func() error {
		return job.onWork(context.Background())
	}

	return &job
}

func (w *Worker) onWork(ctx context.Context) error {
	pendingTransfers, e := w.TransferStore.Top(ctx, 100)
	if e != nil {
		return e
	}

	for _, transfer := range pendingTransfers {
		w.doTransfer(ctx, transfer)
	}

	return nil
}

func (w *Worker) doTransfer(ctx context.Context, transfer *core.Transfer) error {
	return w.DB.Tx(func(tx *db.DB) error {
		input := mixin.TransferInput{
			AssetID:    transfer.AssetID,
			OpponentID: transfer.OpponentID,
			Amount:     transfer.Amount,
			TraceID:    transfer.TraceID,
		}

		input.Memo = transfer.Memo
		if _, e := w.MainWallet.Client.Transfer(ctx, &input, w.MainWallet.Pin); e != nil {
			return e
		}
		//delete record
		if e := w.TransferStore.Delete(ctx, tx, transfer.ID); e != nil {
			return e
		}

		return nil
	})
}
