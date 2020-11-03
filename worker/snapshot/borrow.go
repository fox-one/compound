package snapshot

import (
	"compound/core"
	"context"
)

var handleBorrowEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	_, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID, "")
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	return nil
}
