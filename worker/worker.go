package worker

import (
	"context"
	"time"
)

// Worker worker interface
type Worker interface {
	Run(ctx context.Context) error
}

// TickWorker base worker
type TickWorker struct {
	Delay    time.Duration
	ErrDelay time.Duration
}

// StartTick start tick engine
func (w *TickWorker) StartTick(ctx context.Context, onTick func(ctx context.Context) error) error {
	dur := time.Millisecond

	if w.Delay <= 0 {
		w.Delay = 100 * time.Millisecond
	}

	if w.ErrDelay <= 0 {
		w.ErrDelay = 300 * time.Millisecond
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := onTick(ctx); err == nil {
				dur = w.Delay
			} else {
				dur = w.ErrDelay
			}
		}
	}
}
