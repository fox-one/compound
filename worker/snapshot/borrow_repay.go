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

// from user, refund if error
var handleBorrowRepayEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return w.db.Tx(func(tx *db.DB) error {
		repayAmount := snapshot.Amount.Abs()
		userID := snapshot.OpponentID

		market, e := w.marketStore.Find(ctx, snapshot.AssetID)
		if e != nil {
			return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
		}

		//update interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			return e
		}

		borrow, e := w.borrowStore.Find(ctx, userID, market.Symbol)
		if e != nil {
			return handleRefundEvent(ctx, w, action, snapshot, core.ErrBorrowNotFound)
		}

		//update borrow info
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
		if e != nil {
			return e
		}
		redundantAmount := repayAmount.Sub(borrowBalance)
		newBalance := borrowBalance.Sub(repayAmount)
		newIndex := market.BorrowIndex
		if newBalance.LessThanOrEqual(decimal.Zero) {
			newBalance = decimal.Zero
			newIndex = decimal.Zero
		}

		borrow.Principal = newBalance
		borrow.InterestIndex = newIndex
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			return e
		}

		market.TotalBorrows = market.TotalBorrows.Sub(repayAmount).Truncate(8)
		market.TotalCash = market.TotalCash.Add(repayAmount)

		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			return e
		}

		if redundantAmount.GreaterThan(decimal.Zero) {
			refundAmount := redundantAmount.Truncate(8)
			//refund redundant amount to user
			refundTrace := id.UUIDFromString(fmt.Sprintf("repay-refund-%s", snapshot.TraceID))
			input := mixin.TransferInput{
				AssetID:    snapshot.AssetID,
				OpponentID: userID,
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
