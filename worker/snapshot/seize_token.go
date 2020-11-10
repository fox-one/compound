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
var handleSeizeTokenEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	liquidator := snapshot.OpponentID
	user := action[core.ActionKeyUser]
	seizedSymbol := action[core.ActionKeySymbol]

	repayTokens := snapshot.Amount.Abs()

	curBlock, e := w.blockService.CurrentBlock(ctx)
	if e != nil {
		return e
	}

	supplyMarket, e := w.marketStore.Find(ctx, "", seizedSymbol)
	if e != nil {
		return e
	}

	borrowMarket, e := w.marketStore.Find(ctx, snapshot.AssetID, "")
	if e != nil {
		return e
	}

	supply, e := w.supplyStore.Find(ctx, user, seizedSymbol)
	if e != nil {
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, user, borrowMarket.Symbol)
	if e != nil {
		return e
	}

	borrowPrice, e := w.priceService.GetUnderlyingPrice(ctx, borrowMarket.Symbol, curBlock)
	if e != nil {
		return e
	}

	supplyPrice, e := w.priceService.GetUnderlyingPrice(ctx, seizedSymbol, curBlock)
	if e != nil {
		return e
	}
	seizedPrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
	repayValue := repayTokens.Mul(borrowPrice)
	seizedTokens := repayValue.Div(seizedPrice)

	if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, seizedTokens) {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//seize token successful,send seized token to user
	trace := id.UUIDFromString(fmt.Sprintf("seizetoken-transfer-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    supplyMarket.AssetID,
		OpponentID: liquidator,
		Amount:     seizedTokens,
		TraceID:    trace,
	}
	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceSeizeTokenTransfer
		action[core.ActionKeyUser] = user
		action[core.ActionKeySymbol] = borrowMarket.Symbol
		action[core.ActionKeyAmount] = repayTokens.String()

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

// from system
var handleSeizeTokenTransferEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	seizedAmount := snapshot.Amount.Abs()
	seizedAssetID := snapshot.AssetID
	seizedUserID := action[core.ActionKeyUser]
	borrowSymbol := action[core.ActionKeySymbol]
	repayAmount, e := decimal.NewFromString(action[core.ActionKeyAmount])
	if e != nil {
		return e
	}

	return w.db.Tx(func(tx *db.DB) error {
		supplyMarket, e := w.marketStore.Find(ctx, seizedAssetID, "")
		if e != nil {
			return e
		}
		//update supply
		supply, e := w.supplyStore.Find(ctx, seizedUserID, supplyMarket.Symbol)
		if e != nil {
			return e
		}

		seizedCTokens := supply.CTokens.Mul(seizedAmount).Div(supply.Principal)
		supply.Principal = supply.Principal.Sub(seizedAmount).Truncate(8)
		supply.CTokens = supply.CTokens.Sub(seizedCTokens).Truncate(8)
		supply.CollateTokens = supply.CollateTokens.Sub(seizedCTokens).Truncate(8)
		interestChanged := seizedAmount.Mul(supply.InterestBalance.Div(supply.Principal))
		supply.InterestBalance = supply.InterestBalance.Sub(interestChanged).Truncate(8)
		if e = w.supplyStore.Update(ctx, tx, supply); e != nil {
			return e
		}

		//update market ctokens
		supplyMarket.CTokens = supplyMarket.CTokens.Sub(seizedCTokens).Truncate(8)
		if e = w.marketStore.Update(ctx, tx, supplyMarket); e != nil {
			return e
		}

		//update borrow
		borrow, e := w.borrowStore.Find(ctx, seizedUserID, borrowSymbol)
		if e != nil {
			return e
		}
		borrow.Principal = borrow.Principal.Sub(repayAmount).Truncate(8)
		interestChanged = repayAmount.Mul(borrow.InterestBalance).Div(borrow.Principal)
		borrow.InterestBalance = borrow.InterestBalance.Sub(interestChanged).Truncate(8)
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			return e
		}

		return nil
	})
}
