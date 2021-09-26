package cashier

import (
	"compound/core"
	"context"
	"errors"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/fatih/structs"
	"github.com/fox-one/pkg/logger"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Cashier cashier
//
// use output to spend
type Config struct {
	Batch    int   `json:"batch" valid:"required"`
	Capacity int64 `json:"capacity" valid:"required"`
}

func New(
	wallets core.WalletStore,
	walletz core.WalletService,
	system *core.System,
	cfg Config,
) *Cashier {
	if _, err := govalidator.ValidateStruct(cfg); err != nil {
		panic(err)
	}

	w := &Cashier{
		wallets: wallets,
		walletz: walletz,
		system:  system,
		cfg:     cfg,
	}

	return w
}

type Cashier struct {
	wallets core.WalletStore
	walletz core.WalletService
	system  *core.System
	cfg     Config
}

func (w *Cashier) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "cashier")
	ctx = logger.WithContext(ctx, log)

	log.WithFields(structs.Map(w.cfg)).Infoln("start")

	f := w.sync
	if w.cfg.Capacity > 1 {
		f = w.parallel(w.cfg.Capacity)
	}

	dur := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := w.run(ctx, f); err == nil {
				dur = 300 * time.Millisecond
			} else {
				dur = 500 * time.Millisecond
			}
		}
	}
}

func (w *Cashier) run(ctx context.Context, f func(context.Context, []*core.Transfer) error) error {
	log := logger.FromContext(ctx)

	transfers, err := w.wallets.ListTransfers(ctx, core.TransferStatusAssigned, w.cfg.Batch)
	if err != nil {
		log.WithError(err).Errorln("wallets.ListTransfers")
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
	log := logger.FromContext(ctx).WithField("transfer", transfer.TraceID)
	ctx = logger.WithContext(ctx, log)

	outputs, err := w.wallets.ListSpentBy(ctx, transfer.AssetID, transfer.TraceID)
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
	if tx, err := w.walletz.Spend(ctx, outputs, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("walletz.Spend")
		return err
	} else if tx != nil {
		// signature completed, prepare to send this tx to mixin mainnet
		if err := w.wallets.CreateRawTransaction(ctx, tx); err != nil {
			logger.FromContext(ctx).WithError(err).Errorln("wallets.CreateRawTransaction")
			return err
		}
	}

	transfer.Handled = true
	if err := w.wallets.UpdateTransfer(ctx, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.UpdateTransfer")
		return err
	}

	return nil
}
