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

	repayAmount := snapshot.Amount.Abs()

	blockNum, e := w.blockService.GetBlock(ctx, snapshot.CreatedAt)
	if e != nil {
		return e
	}

	supplyMarket, e := w.marketStore.FindBySymbol(ctx, seizedSymbol)
	if e != nil {
		return e
	}

	borrowMarket, e := w.marketStore.Find(ctx, snapshot.AssetID)
	if e != nil {
		return e
	}

	supply, e := w.supplyStore.Find(ctx, user, supplyMarket.CTokenAssetID)
	if e != nil {
		return e
	}

	borrow, e := w.borrowStore.FindByTrace(ctx, action[core.ActionKeyBorrowTrace])
	if e != nil {
		return e
	}

	borrowPrice, e := w.priceService.GetUnderlyingPrice(ctx, borrowMarket.Symbol, blockNum)
	if e != nil {
		return e
	}

	supplyPrice, e := w.priceService.GetUnderlyingPrice(ctx, seizedSymbol, blockNum)
	if e != nil {
		return e
	}
	seizedPrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
	repayValue := repayAmount.Mul(borrowPrice)
	seizedAmount := repayValue.Div(seizedPrice)

	if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, seizedAmount) {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//seize token successful,send seized token to user
	trace := id.UUIDFromString(fmt.Sprintf("seizetoken-transfer-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    supplyMarket.AssetID,
		OpponentID: liquidator,
		Amount:     seizedAmount,
		TraceID:    trace,
	}
	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceSeizeTokenTransfer
		action[core.ActionKeyUser] = user
		action[core.ActionKeySymbol] = borrowMarket.Symbol
		action[core.ActionKeyAmount] = repayAmount.String()
		action[core.ActionKeyBorrowTrace] = borrow.Trace

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
		supplyMarket, e := w.marketStore.Find(ctx, seizedAssetID)
		if e != nil {
			return e
		}

		changedCTokens := seizedAmount.Div(supplyMarket.ExchangeRate)
		//update supply
		supply, e := w.supplyStore.Find(ctx, seizedUserID, supplyMarket.Symbol)
		if e != nil {
			return e
		}

		supply.Collaterals = supply.Collaterals.Sub(changedCTokens)
		if e = w.supplyStore.Update(ctx, tx, supply); e != nil {
			return e
		}

		//update market ctokens
		supplyMarket.TotalCash = supplyMarket.TotalCash.Sub(seizedAmount).Truncate(8)
		supplyMarket.CTokens = supplyMarket.CTokens.Sub(changedCTokens).Truncate(8)
		if e = w.marketStore.Update(ctx, tx, supplyMarket); e != nil {
			return e
		}

		if e = w.marketService.KeppFlywheelMoving(ctx, tx, supplyMarket, snapshot.CreatedAt); e != nil {
			return e
		}

		borrowMarket, e := w.marketStore.FindBySymbol(ctx, borrowSymbol)
		if e != nil {
			return e
		}

		//update borrow
		borrow, e := w.borrowStore.FindByTrace(ctx, action[core.ActionKeyBorrowTrace])
		if e != nil {
			return e
		}
		borrow.Principal = borrow.Principal.Sub(repayAmount).Truncate(8)
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			return e
		}

		newInterest := repayAmount.Mul(borrow.InterestIndex.Sub(decimal.NewFromInt(1)))
		newPrincipal := repayAmount.Sub(newInterest)
		reserves := newInterest.Mul(borrowMarket.ReserveFactor)

		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Sub(newPrincipal)
		borrowMarket.TotalCash = borrowMarket.TotalCash.Add(repayAmount)
		borrowMarket.Reserves = borrowMarket.Reserves.Add(reserves)

		if e = w.marketService.KeppFlywheelMoving(ctx, tx, borrowMarket, snapshot.CreatedAt); e != nil {
			return e
		}

		if e = w.borrowService.UpdateMarketInterestIndex(ctx, tx, borrowMarket, borrowMarket.BlockNumber); e != nil {
			return e
		}

		return nil
	})
}
