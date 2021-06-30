package cashier

import (
	"context"
	"errors"

	"compound/core"
	"compound/worker"

	"github.com/fox-one/pkg/logger"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Cashier cashier
//
// use output to spend
type Cashier struct {
	worker.TickWorker
	walletStore   core.WalletStore
	walletService core.WalletService
	system        *core.System
	cfg           Config
}

type Config struct {
	Batch    int   `json:"batch" valid:"required"`
	Capacity int64 `json:"capacity" valid:"required"`
}

// New new cashier
func New(
	walletStr core.WalletStore,
	walletSrv core.WalletService,
	system *core.System,
	cfg Config,
) *Cashier {
	cashier := Cashier{
		walletStore:   walletStr,
		walletService: walletSrv,
		system:        system,
		cfg:           cfg,
	}

	return &cashier
}

// Run run worker
func (w *Cashier) Run(ctx context.Context) error {
	f := w.sync
	if w.cfg.Capacity > 1 {
		f = w.parallel(w.cfg.Capacity)
	}

	return w.StartTick(ctx, func(ctx context.Context) error {
		return w.onWork(ctx, f)
	})
}

func (w *Cashier) onWork(ctx context.Context, f func(context.Context, []*core.Transfer) error) error {
	log := logger.FromContext(ctx).WithField("worker", "cashier")

	transfers, err := w.walletStore.ListTransfers(ctx, core.TransferStatusAssigned, w.cfg.Batch)
	if err != nil {
		log.WithError(err).Errorln("list transfers")
		return err
	}

	if len(transfers) == 0 {
		return errors.New("EOF")
	}

	return f(ctx, transfers)
}

func (w *Cashier) sync(ctx context.Context, transfers []*core.Transfer) error {
	for _, transfer := range transfers {
		if err := w.handleTransfer(ctx, transfer); err != nil {
			return err
		}
	}

	return nil
}

func (w *Cashier) parallel(capacity int64) func(ctx context.Context, transfers []*core.Transfer) error {
	sem := semaphore.NewWeighted(capacity)

	return func(ctx context.Context, transfers []*core.Transfer) error {
		g := errgroup.Group{}

		for idx := range transfers {
			transfer := transfers[idx]

			if err := sem.Acquire(ctx, 1); err != nil {
				return g.Wait()
			}

			g.Go(func() error {
				defer sem.Release(1)
				return w.handleTransfer(ctx, transfer)
			})
		}

		return g.Wait()
	}
}

func (w *Cashier) handleTransfer(ctx context.Context, transfer *core.Transfer) error {
	log := logger.FromContext(ctx)

	outputs, err := w.walletStore.ListSpentBy(ctx, transfer.AssetID, transfer.TraceID)
	if err != nil {
		log.WithError(err).Errorln("wallets.ListSpentBy")
		return err
	}

	if len(outputs) == 0 {
		log.Errorln("cannot spent transfer with empty outputs")
		return nil
	}

	return w.spend(ctx, outputs, transfer)
}

func (w *Cashier) spend(ctx context.Context, outputs []*core.Output, transfer *core.Transfer) error {
	if tx, err := w.walletService.Spend(ctx, outputs, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("walletz.Spend")
		return err
	} else if tx != nil {
		// signature completed, prepare to send this tx to mixin mainnet
		if err := w.walletStore.CreateRawTransaction(ctx, tx); err != nil {
			logger.FromContext(ctx).WithError(err).Errorln("wallets.CreateRawTransaction")
			return err
		}
	}

	transfer.Handled = true
	if err := w.walletStore.UpdateTransfer(ctx, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.UpdateTransfer")
		return err
	}

	return nil
}
