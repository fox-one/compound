package snapshot

import (
	"compound/core"
	"context"
)

var handleSupplyInterestEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return nil
}

var handleBorrowInterestEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return nil
}
