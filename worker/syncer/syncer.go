package syncer

import (
	"context"
	"errors"

	"compound/core"
	"compound/worker"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
)

const checkpointKey = "sync_checkpoint"

// Syncer sync output
type Syncer struct {
	worker.TickWorker
	walletStore   core.WalletStore
	walletService core.WalletService
	property      property.Store
}

// New new sync worker
func New(walletStr core.WalletStore,
	walletSrv core.WalletService,
	property property.Store,
) *Syncer {
	syncer := Syncer{
		walletStore:   walletStr,
		walletService: walletSrv,
		property:      property,
	}

	return &syncer
}

// Run run worker
func (w *Syncer) Run(ctx context.Context) error {
	return w.StartTick(ctx, func(ctx context.Context) error {
		return w.onWork(ctx)
	})
}

func (w *Syncer) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx)

	v, err := w.property.Get(ctx, checkpointKey)
	if err != nil {
		log.WithError(err).Errorln("property.Get", checkpointKey)
		return err
	}

	offset := v.Time()

	const limit = 500
	outputs, err := w.walletService.Pull(ctx, offset, limit)
	if err != nil {
		log.WithError(err).Errorln("walletz.Pull")
		return err
	}

	if len(outputs) == 0 {
		return errors.New("EOF")
	}

	nextOffset := outputs[len(outputs)-1].UpdatedAt
	end := len(outputs) < limit

	if err := w.walletStore.Save(ctx, outputs, end); err != nil {
		log.WithError(err).Errorln("wallets.Save")
		return err
	}

	if err := w.property.Save(ctx, checkpointKey, nextOffset); err != nil {
		log.WithError(err).Errorln("property.Save:", checkpointKey)
		return err
	}

	return nil
}
