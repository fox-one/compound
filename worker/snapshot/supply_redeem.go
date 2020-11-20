package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/pkg/store/db"
)

// from user
var handleSupplyRedeemEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	ctokenAssetID := snapshot.AssetID
	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		return e
	}

	redeemTokens := snapshot.Amount.Abs()

	// check redeem allowed
	allowed := w.supplyService.RedeemAllowed(ctx, redeemTokens, market)
	if !allowed {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrRedeemNotAllowed)
	}

	// transfer asset to user
	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return e
	}

	amount := redeemTokens.Mul(exchangeRate).Truncate(8)

	return w.db.Tx(func(tx *db.DB) error {
		market.TotalCash = market.TotalCash.Sub(amount)
		market.CTokens = market.CTokens.Sub(redeemTokens)
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			return e
		}

		//transfer to user
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceRedeemTransfer
		memoStr, e := memo.Format()
		if e != nil {
			return e
		}

		trace := id.UUIDFromString(fmt.Sprintf("redeem:%s", snapshot.TraceID))
		transfer := core.Transfer{
			AssetID:    market.AssetID,
			OpponentID: snapshot.OpponentID,
			Amount:     amount,
			TraceID:    trace,
			Memo:       memoStr,
		}
		if e = w.transferStore.Create(ctx, tx, &transfer); e != nil {
			return e
		}

		return nil
	})
}
