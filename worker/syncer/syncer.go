package syncer

import (
	"context"
	"errors"

	"compound/core"
	"compound/worker"

	"compound/internal/mixinet"

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

	var (
		outputs   = make([]*core.Output, 0, 8)
		positions = make(map[string]int)
		pos       = 0
	)

	const Limit = 500

	for {
		batch, err := w.walletService.Pull(ctx, offset, Limit)
		if err != nil {
			log.WithError(err).Errorln("walletz.Pull")
			return err
		}
		for _, u := range batch {
			offset = u.UpdatedAt

			p, ok := positions[u.TraceID]
			if ok {
				outputs[p] = u
				continue
			}

			outputs = append(outputs, u)
			positions[u.TraceID] = pos
			pos++
		}

		if len(batch) < Limit {
			break
		}
	}

	if len(outputs) == 0 {
		return errors.New("EOF")
	}

	mixinet.SortOutputs(outputs)
	if err := w.walletStore.Save(ctx, outputs); err != nil {
		log.WithError(err).Errorln("wallets.Save------------------------------")
		return err
	}

	log.Infoln("save output successful")

	if err := w.property.Save(ctx, checkpointKey, offset); err != nil {
		log.WithError(err).Errorln("property.Save---------------------------------", checkpointKey)
		return err
	}

	log.Infoln("save property successful")

	return nil
}
