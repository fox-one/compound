package block

import (
	"context"
	"time"

	"github.com/fox-one/pkg/logger"
)

//BlockWorker block worker
type BlockWorker struct {
}

// New new block worker
func New() *BlockWorker {
	return &BlockWorker{}
}

// Run block worker run
func (w *BlockWorker) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "block")
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

func (w *BlockWorker) realRun(ctx context.Context) error {
	//TODO calculate blocks,
	// calculate supplyRate
	// calculate bollowRate
	// calculate liquidatable account
	//
	return nil
}
