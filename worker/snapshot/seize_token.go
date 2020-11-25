package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from user
var handleSeizeTokenEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	liquidator := snapshot.OpponentID
	userID := action[core.ActionKeyUser]
	seizedSymbol := action[core.ActionKeySymbol]

	repayAmount := snapshot.Amount.Abs()

	supplyMarket, e := w.marketStore.FindBySymbol(ctx, seizedSymbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	borrowMarket, e := w.marketStore.Find(ctx, snapshot.AssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	//supply market accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, supplyMarket, snapshot.CreatedAt); e != nil {
		return e
	}

	//borrow market accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, borrowMarket, snapshot.CreatedAt); e != nil {
		return e
	}

	supply, e := w.supplyStore.Find(ctx, userID, supplyMarket.CTokenAssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrSupplyNotFound)
	}

	borrow, e := w.borrowStore.Find(ctx, userID, borrowMarket.AssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrBorrowNotFound)
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
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrSeizeNotAllowed)
	}

	return w.db.Tx(func(tx *db.DB) error {
		exchangeRate, e := w.marketService.CurExchangeRate(ctx, supplyMarket)
		if e != nil {
			return e
		}

		changedCTokens := seizedAmount.Div(exchangeRate)
		//update supply
		supply, e := w.supplyStore.Find(ctx, userID, supplyMarket.CTokenAssetID)
		if e != nil {
			return e
		}

		supply.Collaterals = supply.Collaterals.Sub(changedCTokens)
		if e = w.supplyStore.Update(ctx, tx, supply); e != nil {
			return e
		}

		//update supply market ctokens
		supplyMarket.TotalCash = supplyMarket.TotalCash.Sub(seizedAmount).Truncate(8)
		supplyMarket.CTokens = supplyMarket.CTokens.Sub(changedCTokens).Truncate(8)
		if e = w.marketStore.Update(ctx, tx, supplyMarket); e != nil {
			return e
		}

		// update borrow account and borrow market
		borrow, e := w.borrowStore.Find(ctx, userID, borrowMarket.AssetID)
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

		//transfer seized asset to user
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceSeizeTokenTransfer

		memoStr, e := action.Format()
		if e != nil {
			return e
		}
		trace := id.UUIDFromString(fmt.Sprintf("seizetoken-transfer-%s", snapshot.TraceID))
		transfer := core.Transfer{
			AssetID:    supplyMarket.AssetID,
			OpponentID: liquidator,
			Amount:     seizedAmount,
			TraceID:    trace,
			Memo:       memoStr,
		}
		if e = w.transferStore.Create(ctx, tx, &transfer); e != nil {
			return e
		}

		//refund redundant assets to liquidator
		if redundantAmount.GreaterThan(decimal.Zero) {
			refundAmount := redundantAmount.Truncate(8)

			action := core.NewAction()
			action[core.ActionKeyService] = core.ActionServiceRefund
			memoStr, e := action.Format()
			if e != nil {
				return e
			}
			refundTrace := id.UUIDFromString(fmt.Sprintf("liquidate-refund-%s", snapshot.TraceID))
			transfer := core.Transfer{
				AssetID:    snapshot.AssetID,
				OpponentID: snapshot.OpponentID,
				Amount:     refundAmount,
				TraceID:    refundTrace,
				Memo:       memoStr,
			}

			if e = w.transferStore.Create(ctx, tx, &transfer); e != nil {
				return e
			}
		}

		return nil
	})
}
