package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

var handleSupplyRedeemEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	market, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID, "")
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	// redeemTokens := snapshot.Amount.Abs()

	// supplies, e := w.supplyStore.Find(ctx, snapshot.OpponentID, market.Symbol)
	// if e != nil {
	// 	return e
	// }

	//update market

	//update user supply account

	// transfer asset to user
	//mint ctoken
	memo := make(core.Action)
	memo[core.ActionKeyService] = core.ActionServiceRedeemTransfer
	memo[core.ActionKeyCToken] = snapshot.Amount.Abs().String()
	memoStr, e := w.blockService.FormatBlockMemo(ctx, memo)
	if e != nil {
		return e
	}

	trace := id.UUIDFromString(fmt.Sprintf("redeem:%s", snapshot.TraceID))
	_, e = w.dapp.Transfer(ctx, &mixin.TransferInput{
		AssetID:    market.AssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     decimal.Zero,
		TraceID:    trace,
		Memo:       memoStr,
	}, w.config.Mixin.Pin)

	if e != nil {
		return e
	}

	return nil

	return nil
}
