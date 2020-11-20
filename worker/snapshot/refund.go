package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

var handleRefundEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot, errCode core.ErrorCode) error {
	if snapshot.Amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	action = core.NewAction()
	action[core.ActionKeyService] = core.ActionServiceRefund
	action[core.ActionKeyReferTrace] = snapshot.TraceID
	action[core.ActionKeyErrorCode] = errCode.String()
	memoStr, e := action.Format()
	if e != nil {
		return e
	}

	trace := id.UUIDFromString(fmt.Sprintf("refund-%s", snapshot.TraceID))
	transfer := core.Transfer{
		AssetID:    snapshot.AssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     snapshot.Amount.Abs(),
		TraceID:    trace,
		Memo:       memoStr,
	}

	if e := w.transferStore.Create(ctx, w.db, &transfer); e != nil {
		return e
	}

	return nil
}
