package txsender

import (
	"context"
	"errors"
	"time"

	"compound/core"
	"compound/worker"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"
)

// Sender tx sender
type Sender struct {
	worker.BaseJob
	wallets core.WalletStore
}

func New(location string, wallets core.WalletStore) *Sender {
	sender := Sender{
		wallets: wallets,
	}

	l, _ := time.LoadLocation(location)
	sender.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 100ms"
	sender.Cron.AddFunc(spec, sender.Run)
	sender.OnWork = func() error {
		return sender.onWork(context.Background())
	}

	return &sender
}

func (w *Sender) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "txsender")
	const Limit = 20

	txs, err := w.wallets.ListPendingRawTransactions(ctx, Limit)
	if err != nil {
		log.WithError(err).Errorln("list raw transactions")
		return err
	}

	if len(txs) == 0 {
		return errors.New("EOF")
	}

	var g errgroup.Group
	for _, tx := range txs {
		tx := tx
		g.Go(func() error {
			return w.handleRawTransaction(ctx, tx)
		})
	}

	return g.Wait()
}

func (w *Sender) handleRawTransaction(ctx context.Context, tx *core.RawTransaction) error {
	log := logger.FromContext(ctx).WithField("trace_id", tx.TraceID)
	ctx = logger.WithContext(ctx, log)

	if err := w.submitRawTransaction(ctx, tx.Data); err != nil {
		return err
	}
	if err := w.wallets.ExpireRawTransaction(ctx, tx); err != nil {
		log.WithError(err).Errorln("wallets.ExpireRawTransaction")
		return err
	}
	return nil
}

func (w *Sender) submitRawTransaction(ctx context.Context, raw string) error {
	log := logger.FromContext(ctx)
	ctx = mixin.WithMixinNetHost(ctx, mixin.RandomMixinNetHost())

	if tx, err := mixin.SendRawTransaction(ctx, raw); err != nil {
		log.WithError(err).Errorln("SendRawTransaction failed")
		return err
	} else if tx.Snapshot != nil {
		return nil
	}

	var txHash mixin.Hash
	if tx, err := mixin.TransactionFromRaw(raw); err == nil {
		txHash, _ = tx.TransactionHash()
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	dur := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return errors.New("mixin net snapshot not generated")
		case <-time.After(dur):
			if tx, err := mixin.GetTransaction(ctx, txHash); err != nil {
				log.WithError(err).Errorln("GetTransaction failed")
				return err
			} else if tx.Snapshot != nil {
				return nil
			}
			dur = time.Second
		}
	}
}
