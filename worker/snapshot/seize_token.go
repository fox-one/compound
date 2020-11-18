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

	supplyMarket, e := w.marketStore.FindBySymbol(ctx, seizedSymbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	borrowMarket, e := w.marketStore.Find(ctx, snapshot.AssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//supply market accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, supplyMarket, snapshot.CreatedAt); e != nil {
		return e
	}

	//borrow market accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, borrowMarket, snapshot.CreatedAt); e != nil {
		return e
	}

	supply, e := w.supplyStore.Find(ctx, user, supplyMarket.CTokenAssetID)
	if e != nil {
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, user, borrowMarket.Symbol)
	if e != nil {
		return e
	}

	borrowPrice, e := w.priceService.GetCurrentUnderlyingPrice(ctx, borrowMarket)
	if e != nil {
		return e
	}

	supplyPrice, e := w.priceService.GetCurrentUnderlyingPrice(ctx, supplyMarket)
	if e != nil {
		return e
	}
	seizedPrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
	repayValue := repayAmount.Mul(borrowPrice)
	seizedAmount := repayValue.Div(seizedPrice)

	// refund to liquidator if seize not allowed
	if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, seizedAmount, snapshot.CreatedAt) {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//seize token successful,send seized token to liquidator
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

		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, supplyMarket, snapshot.CreatedAt); e != nil {
			return e
		}

		borrowMarket, e := w.marketStore.FindBySymbol(ctx, borrowSymbol)
		if e != nil {
			return e
		}

		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, borrowMarket, snapshot.CreatedAt); e != nil {
			return e
		}

		exchangeRate, e := w.marketService.CurExchangeRate(ctx, supplyMarket)
		if e != nil {
			return e
		}

		changedCTokens := seizedAmount.Div(exchangeRate)
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

		// update borrow account and borrow market
		borrow, e := w.borrowStore.Find(ctx, seizedUserID, borrowSymbol)
		if e != nil {
			return e
		}
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, borrowMarket)
		if e != nil {
			return e
		}
		redundantAmount := repayAmount.Sub(borrowBalance)
		newBorrowBalance := borrowBalance.Sub(repayAmount).Truncate(8)
		newIndex := borrowMarket.BorrowIndex
		if newBorrowBalance.LessThanOrEqual(decimal.Zero) {
			newBorrowBalance = decimal.Zero
			newIndex = decimal.Zero
		}
		borrow.Principal = newBorrowBalance
		borrow.InterestIndex = newIndex
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			return e
		}

		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Sub(repayAmount)
		borrowMarket.TotalCash = borrowMarket.TotalCash.Add(repayAmount)
		if e = w.marketStore.Update(ctx, tx, borrowMarket); e != nil {
			return e
		}

		if redundantAmount.GreaterThan(decimal.Zero) {
			//refund redundant assets to liquidator
			refundAmount := redundantAmount.Truncate(8)

			refundTrace := id.UUIDFromString(fmt.Sprintf("liquidate-refund-%s", snapshot.TraceID))
			input := mixin.TransferInput{
				AssetID:    snapshot.AssetID,
				OpponentID: snapshot.OpponentID,
				Amount:     refundAmount,
				TraceID:    refundTrace,
			}

			if !w.walletService.VerifyPayment(ctx, &input) {
				action := core.NewAction()
				action[core.ActionKeyService] = core.ActionServiceRefund
				memoStr, e := action.Format()
				if e != nil {
					return e
				}
				input.Memo = memoStr
				if _, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin); e != nil {
					return e
				}
			}
		}

		return nil
	})
}
