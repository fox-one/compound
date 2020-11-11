package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
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
		memoStr, e := w.blockService.FormatBlockMemo(ctx, memo)
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
	tokens := snapshot.Amount
	userID := snapshot.OpponentID
	ctokenAssetID := snapshot.AssetID

	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID, "")
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	supply, e := w.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	remainTokens := supply.CTokens.Sub(supply.CollateTokens)
	if tokens.GreaterThan(remainTokens) {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//update supply
	supply.CollateTokens = supply.CollateTokens.Add(tokens)
	e = w.supplyStore.Update(ctx, w.db, supply)
	if e != nil {
		return e
	}

	return nil
}

// from system，ignored if error
var handleUnpledgeEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "unpledge")

	unpledgedTokens, e := decimal.NewFromString(action[core.ActionKeyCToken])
	if e != nil {
		return nil
	}
	symbol := action[core.ActionKeySymbol]
	userID := action[core.ActionKeyUser]

	market, e := w.marketStore.Find(ctx, "", symbol)
	if e != nil {
		return nil
	}

	supply, e := w.supplyStore.Find(ctx, userID, symbol)
	if e != nil {
		return nil
	}

	if unpledgedTokens.GreaterThanOrEqual(supply.CollateTokens) {
		return nil
	}

	liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		return nil
	}

	curBlock, e := w.blockService.CurrentBlock(ctx)
	if e != nil {
		return nil
	}

	price, e := w.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
	if e != nil {
		return nil
	}

	//calculate unpledge token liquidity
	unpledgedTokenLiquidity := unpledgedTokens.Mul(supply.Principal).Div(supply.CTokens).Mul(market.CollateralFactor).Mul(price)
	if unpledgedTokenLiquidity.GreaterThanOrEqual(liquidity) {
		log.Errorln("insufficient liquidity")
		return nil
	}

	trace := id.UUIDFromString(fmt.Sprintf("unpledge-transfer-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    market.CTokenAssetID,
		Amount:     unpledgedTokens,
		OpponentID: userID,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceUnpledgeTransfer
		action[core.ActionKeySymbol] = symbol
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

// send unpledge ctoken to user
var handleUnpledgeTransferEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := snapshot.OpponentID
	unpledgedTokens := snapshot.Amount.Abs()
	symbol := action[core.ActionKeySymbol]

	supply, e := w.supplyStore.Find(ctx, userID, symbol)
	if e != nil {
		return nil
	}

	//update supply
	supply.CollateTokens = supply.CollateTokens.Sub(unpledgedTokens)
	e = w.supplyStore.Update(ctx, w.db, supply)
	if e != nil {
		return e
	}
	return nil
}
