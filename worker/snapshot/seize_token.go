package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"
	"strings"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from user
var handleSeizeTokenEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	liquidator := snapshot.OpponentID
	userID := action[core.ActionKeyUser]
	seizedSymbol := strings.ToUpper(action[core.ActionKeySymbol])

	userPayAmount := snapshot.Amount.Abs()

	// to seize
	supplyMarket, e := w.marketStore.FindBySymbol(ctx, seizedSymbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	supplyExchangeRate, e := w.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		return e
	}

	// to repay
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

	if borrowPrice.LessThanOrEqual(decimal.Zero) {
		return e
	}

	supplyPrice, e := w.priceService.GetCurrentUnderlyingPrice(ctx, supplyMarket)
	if e != nil {
		return e
	}
	if supplyPrice.LessThanOrEqual(decimal.Zero) {
		return e
	}

	// refund to liquidator if seize not allowed
	if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, snapshot.CreatedAt) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrSeizeNotAllowed)
	}

	return w.db.Tx(func(tx *db.DB) error {
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, borrowMarket)
		if e != nil {
			return e
		}

		maxSeize := supply.Collaterals.Mul(supplyExchangeRate).Mul(supplyMarket.CloseFactor)
		seizedPrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
		maxSeizeValue := maxSeize.Mul(seizedPrice)
		repayValue := userPayAmount.Mul(borrowPrice)
		borrowBalanceValue := borrowBalance.Mul(borrowPrice)
		seizedAmount := repayValue.Div(seizedPrice)
		if repayValue.GreaterThan(maxSeizeValue) {
			repayValue = maxSeizeValue
			seizedAmount = repayValue.Div(seizedPrice)
		}

		if repayValue.GreaterThan(borrowBalanceValue) {
			repayValue = borrowBalanceValue
			seizedAmount = repayValue.Div(seizedPrice)
		}

		seizedCTokens := seizedAmount.Div(supplyExchangeRate)
		//update supply
		supply.Collaterals = supply.Collaterals.Sub(seizedCTokens)
		if e = w.supplyStore.Update(ctx, tx, supply); e != nil {
			return e
		}

		//update supply market ctokens
		supplyMarket.TotalCash = supplyMarket.TotalCash.Sub(seizedAmount).Truncate(8)
		supplyMarket.CTokens = supplyMarket.CTokens.Sub(seizedCTokens).Truncate(8)
		if e = w.marketStore.Update(ctx, tx, supplyMarket); e != nil {
			return e
		}

		// update borrow account and borrow market
		repayAmount := repayValue.Div(borrowPrice)
		redundantAmount := userPayAmount.Sub(repayAmount)
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
				OpponentID: liquidator,
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
