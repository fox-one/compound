package snapshot

import (
	"compound/core"
	"context"
)

func (w *Worker) handleBorrowRepayEvent(ctx context.Context, snapshot *core.Snapshot) error {
	_, e := w.marketStore.Find(ctx, snapshot.AssetID, "")
	if e != nil {
		return w.handleRefundEvent(ctx, snapshot)
	}

	return nil
}
