package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

// from user
var handleSupplyEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	market, e := w.marketStore.Find(ctx, snapshot.AssetID)
	if e != nil {
		//refund to user
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		return e
	}
	ctokens := snapshot.Amount.Div(exchangeRate).Truncate(8)

	trace := id.UUIDFromString(fmt.Sprintf("mint:%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    market.CTokenAssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     ctokens,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		//mint ctoken
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceMint
		memo[core.ActionKeyAmount] = snapshot.Amount.Abs().String()
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

// from user, refund if error
var handlePledgeEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	ctokens := snapshot.Amount
	userID := snapshot.OpponentID
	ctokenAssetID := snapshot.AssetID

	supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
	if e != nil {
		//new
		supply = &core.Supply{
			UserID:        userID,
			CTokenAssetID: ctokenAssetID,
			Collaterals:   ctokens,
		}
		if e = w.supplyStore.Save(ctx, w.db, supply); e != nil {
			return e
		}
	} else {
		//update supply
		supply.Collaterals = supply.Collaterals.Add(ctokens)
		e = w.supplyStore.Update(ctx, w.db, supply)
		if e != nil {
			return e
		}
	}

	return nil
}

// from system
var handleUnpledgeEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := snapshot.OpponentID
	unpledgedTokens := snapshot.Amount.Abs()
	ctokenAssetID := snapshot.AssetID

	supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
	if e != nil {
		return e
	}

	//update supply
	supply.Collaterals = supply.Collaterals.Sub(unpledgedTokens)
	if supply.Collaterals.LessThan(decimal.Zero) {
		supply.Collaterals = decimal.Zero
	}
	if e = w.supplyStore.Update(ctx, w.db, supply); e != nil {
		return e
	}

	return nil
}
