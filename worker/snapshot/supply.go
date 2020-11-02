package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
)

func (w *Worker) handleSupplyEvent(ctx context.Context, snapshot *core.Snapshot) error {
	market, e := w.marketStore.Find(ctx, snapshot.AssetID, "")
	if e != nil {
		return w.handleRefundEvent(ctx, snapshot)
	}

	//update market
	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return e
	}
	ctokens := snapshot.Amount.Div(exchangeRate)

	//mint ctoken
	memo := make(core.Action)
	memo[core.ActionKeyService] = core.ActionServiceMint
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
