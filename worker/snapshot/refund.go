package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

var handleRefundEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	if snapshot.Amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	trace := id.UUIDFromString(fmt.Sprintf("refund-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    snapshot.AssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     snapshot.Amount.Abs(),
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceRefund
		memoStr, e := action.Format()
		if e != nil {
			return e
		}

		input.Memo = memoStr
		if _, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin); e != nil {
			return e
		}
	}

	return nil
}
