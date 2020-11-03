package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"
	"strconv"

	"github.com/fox-one/mixin-sdk-go"
)

var handleSupplyEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	market, e := w.marketStore.Find(ctx, snapshot.AssetID, "")
	if e != nil {
		//refund to user
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return e
	}
	ctokens := snapshot.Amount.Div(exchangeRate).Truncate(8)

	//mint ctoken
	memo := make(core.Action)
	memo[core.ActionKeyService] = core.ActionServiceMint
	memo[core.ActionKeyAmount] = snapshot.Amount.Abs().String()
	memoStr, e := w.blockService.FormatBlockMemo(ctx, memo)
	if e != nil {
		return e
	}
	trace := id.UUIDFromString(fmt.Sprintf("mint:%s", snapshot.TraceID))
	_, e = w.dapp.Transfer(ctx, &mixin.TransferInput{
		AssetID:    market.CTokenAssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     ctokens,
		TraceID:    trace,
		Memo:       memoStr,
	}, w.config.Mixin.Pin)

	if e != nil {
		return e
	}

	return nil
}

var handleUpdateCollateralStatusEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	user := action[core.ActionKeyUser]
	symbol := action[core.ActionKeySymbol]
	statusStr := action[core.ActionKeyStatus]
	status, _ := strconv.Atoi(statusStr)

	suppy, e := w.supplyStore.Find(ctx, user, symbol)
	if e != nil {
		return nil
	}

	suppy.CollateStatus = core.CollateStatus(status)
	return w.supplyStore.Update(ctx, w.db, suppy)
}
