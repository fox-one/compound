package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
)

var handleRequestAccountLiquidityEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	if snapshot.AssetID != w.config.App.GasAssetID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	userID := action[core.ActionKeyUser]

	blockNum, e := w.blockService.GetBlock(ctx, snapshot.CreatedAt)
	if e != nil {
		return e
	}

	liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID, blockNum)
	if e != nil {
		return e
	}

	trace := id.UUIDFromString(fmt.Sprintf("liquidity-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    w.config.App.GasAssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     core.GasCost,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceLiquidityResponse
		action[core.ActionKeyUser] = userID
		action[core.ActionKeyLiquidity] = liquidity.Truncate(8).String()
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
