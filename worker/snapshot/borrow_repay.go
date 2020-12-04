package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from user, refund if error
var handleBorrowRepayEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "borrow_repay")

	repayAmount := snapshot.Amount.Abs()
	userID := snapshot.OpponentID

	market, e := w.marketStore.Find(ctx, snapshot.AssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	//update interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, market.AssetID)
	if e != nil {
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrBorrowNotFound)
	}

	return w.db.Tx(func(tx *db.DB) error {
		//update borrow info
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
		if e != nil {
			log.Errorln(e)
			return e
		}
		realRepaidBalance := repayAmount
		redundantAmount := repayAmount.Sub(borrowBalance)
		newBalance := borrowBalance.Sub(repayAmount)
		newIndex := market.BorrowIndex
		if newBalance.LessThanOrEqual(decimal.Zero) {
			newBalance = decimal.Zero
			newIndex = decimal.Zero
			realRepaidBalance = borrowBalance
		}

		borrow.Principal = newBalance
		borrow.InterestIndex = newIndex
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			log.Errorln(e)
			return e
		}

		market.TotalBorrows = market.TotalBorrows.Sub(realRepaidBalance).Truncate(8)
		market.TotalCash = market.TotalCash.Add(realRepaidBalance)

		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		//update interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		if redundantAmount.GreaterThan(decimal.Zero) {
			refundAmount := redundantAmount.Truncate(8)
			//refund redundant amount to user
			action := core.NewAction()
			action[core.ActionKeyService] = core.ActionServiceRefund
			memoStr, e := action.Format()
			if e != nil {
				log.Errorln(e)
				return e
			}
			refundTrace := id.UUIDFromString(fmt.Sprintf("repay-refund-%s", snapshot.TraceID))
			input := core.Transfer{
				AssetID:    snapshot.AssetID,
				OpponentID: userID,
				Amount:     refundAmount,
				TraceID:    refundTrace,
				Memo:       memoStr,
			}

			if e = w.transferStore.Create(ctx, tx, &input); e != nil {
				log.Errorln(e)
				return e
			}
		}

		return nil
	})
}
