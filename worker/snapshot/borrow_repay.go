package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleReplayEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "borrow_repay")

	repayAmount := output.Amount
	assetID := output.AssetID

	market, e := w.marketStore.Find(ctx, assetID)
	if e != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}

	//update interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, output.UpdatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, market.AssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrBorrowNotFound, "")
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

		borrow.Principal = newBalance.Truncate(16)
		borrow.InterestIndex = newIndex.Truncate(16)
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			log.Errorln(e)
			return e
		}

		market.TotalBorrows = market.TotalBorrows.Sub(realRepaidBalance).Truncate(16)
		market.TotalCash = market.TotalCash.Add(realRepaidBalance).Truncate(16)

		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		//update interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, output.UpdatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		if redundantAmount.GreaterThan(decimal.Zero) {
			refundAmount := redundantAmount.Truncate(8)
			transferAction := core.TransferAction{
				Source:        core.ActionTypeRefundTransfer,
				TransactionID: followID,
			}

			return w.transferOut(ctx, userID, followID, output.TraceID, assetID, refundAmount, &transferAction)
		}

		return nil
	})
}
