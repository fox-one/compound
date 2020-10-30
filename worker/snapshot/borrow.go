package snapshot

import (
	"compound/core"
	"context"
)

func (w *Worker) handleBorrowEvent(ctx context.Context, snapshot *core.Snapshot) error {
	_, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID, "")
	if e != nil {
		return w.handleRefundEvent(ctx, snapshot)
	}

	return nil
}
