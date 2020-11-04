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
	market, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID, "")
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	supply, e := w.supplyStore.Find(ctx, snapshot.OpponentID, market.Symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	redeemTokens := snapshot.Amount.Abs()

	// check redeem allowed
	allowed := w.supplyService.RedeemAllowed(ctx, redeemTokens, snapshot.OpponentID, market)
	if !allowed {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	// transfer asset to user
	amount := redeemTokens.Mul(supply.Principal).Div(supply.Ctokens)
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
		memoStr, e := w.blockService.FormatBlockMemo(ctx, memo)
		if e != nil {
			return e
		}

		input.Memo = memoStr
		_, e = w.dapp.Transfer(ctx, &input, w.config.Mixin.Pin)

		if e != nil {
			return e
		}
	}

	return nil
}

//redeem transfer callback, to user
var handleRedeemTransferEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := snapshot.OpponentID
	assetID := snapshot.AssetID
	amount := snapshot.Amount

	reducedCtokens, e := decimal.NewFromString(action[core.ActionKeyCToken])
	if e != nil {
		return e
	}

	return w.db.Tx(func(tx *db.DB) error {
		//update market
		market, e := w.marketStore.Find(ctx, assetID, "")
		if e != nil {
			return e
		}
		market.CTokens = market.CTokens.Sub(reducedCtokens)
		e = w.marketStore.Update(ctx, tx, market)
		if e != nil {
			return e
		}

		//update user supply account
		supply, e := w.supplyStore.Find(ctx, userID, assetID)
		if e != nil {
			return e
		}
		supply.Ctokens = supply.Ctokens.Sub(reducedCtokens)
		supply.Principal = supply.Principal.Sub(amount)
		if supply.Ctokens.LessThanOrEqual(decimal.Zero) {
			supply.Ctokens = decimal.Zero
			supply.Principal = decimal.Zero
		}
		e = w.supplyStore.Update(ctx, tx, supply)
		if e != nil {
			return e
		}

		return nil
	})
}
