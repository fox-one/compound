package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"errors"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
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
	log := logger.FromContext(ctx).WithField("worker", "unpledge")

	userID := action[core.ActionKeyUser]
	symbol := action[core.ActionKeySymbol]
	unpledgedTokens, e := decimal.NewFromString(action[core.ActionKeyCToken])
	if e != nil {
		log.Errorln(e)
		return nil
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		log.Errorln(e)
		return nil
	}

	supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		log.Errorln(e)
		return nil
	}

	if unpledgedTokens.GreaterThanOrEqual(supply.Collaterals) {
		log.Errorln(errors.New("insufficient collaterals"))
		return nil
	}

	liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		log.Errorln(e)
		return nil
	}

	blockNum, e := w.blockService.GetBlock(ctx, snapshot.CreatedAt)
	if e != nil {
		log.Errorln(e)
		return nil
	}

	price, e := w.priceService.GetUnderlyingPrice(ctx, market.Symbol, blockNum)
	if e != nil {
		log.Errorln(e)
		return nil
	}

	exchangeRate := market.ExchangeRate
	unpledgedTokenLiquidity := unpledgedTokens.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
	if unpledgedTokenLiquidity.GreaterThanOrEqual(liquidity) {
		log.Errorln(errors.New("insufficient liquidity"))
		return nil
	}

	trace := id.UUIDFromString(fmt.Sprintf("unpledge-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    market.CTokenAssetID,
		OpponentID: userID,
		Amount:     unpledgedTokens,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceUnpledgeTransfer

		memoStr, e := action.Format()
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

var handleUnpledgeTransferEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := snapshot.OpponentID
	ctokenAssetID := snapshot.AssetID
	unpledgedCTokens := snapshot.Amount.Abs()

	supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
	if e != nil {
		return e
	}

	supply.Collaterals = supply.Collaterals.Sub(unpledgedCTokens)
	if supply.Collaterals.LessThan(decimal.Zero) {
		supply.Collaterals = decimal.Zero
	}

	if e = w.supplyStore.Update(ctx, w.db, supply); e != nil {
		return e
	}

	return nil
}
