package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from user
var handleSupplyRedeemEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	ctokenAssetID := snapshot.AssetID
	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		return e
	}

	redeemTokens := snapshot.Amount.Abs()

	// check redeem allowed
	allowed := w.supplyService.RedeemAllowed(ctx, redeemTokens, market)
	if !allowed {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	// transfer asset to user
	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	amount := redeemTokens.Mul(exchangeRate).Truncate(8)
	trace := id.UUIDFromString(fmt.Sprintf("redeem:%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    market.AssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     amount,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceRedeemTransfer
		memo[core.ActionKeyCToken] = snapshot.Amount.Abs().String()
		memoStr, e := memo.Format()
		if e != nil {
			return e
		}

		input.Memo = memoStr
		_, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin)

		if e != nil {
			return e
		}
	}

	return nil
}

//redeem transfer callback, to user
var handleRedeemTransferEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return w.db.Tx(func(tx *db.DB) error {
		assetID := snapshot.AssetID
		amount := snapshot.Amount.Abs()

		redeemedCtokens, e := decimal.NewFromString(action[core.ActionKeyCToken])
		if e != nil {
			return e
		}

		market, e := w.marketStore.Find(ctx, assetID)
		if e != nil {
			return e
		}
		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			return e
		}

		market.TotalCash = market.TotalCash.Sub(amount)
		market.CTokens = market.CTokens.Sub(redeemedCtokens)
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			return e
		}

		return nil
	})
}
