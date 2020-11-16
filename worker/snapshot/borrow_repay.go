package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from user, refund if error
var handleBorrowRepayEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return w.db.Tx(func(tx *db.DB) error {
		repayAmount := snapshot.Amount.Abs()
		borrowTrace := action[core.ActionKeyBorrowTrace]

		market, e := w.marketStore.Find(ctx, snapshot.AssetID)
		if e != nil {
			return handleRefundEvent(ctx, w, action, snapshot)
		}

		borrow, e := w.borrowStore.FindByTrace(ctx, borrowTrace)
		if e != nil {
			return handleRefundEvent(ctx, w, action, snapshot)
		}

		curInterest := repayAmount.Mul(borrow.InterestIndex.Sub(decimal.NewFromInt(1)))
		newPrincipal := borrow.Principal.Sub(repayAmount.Sub(curInterest))
		reserves := curInterest.Mul(market.ReserveFactor)

		//update borrow info

		borrow.Principal = borrow.Principal.Sub(newPrincipal)
		if borrow.Principal.LessThan(decimal.Zero) {
			borrow.Principal = decimal.Zero
		}
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			return e
		}

		market.TotalBorrows = market.TotalBorrows.Sub(newPrincipal).Truncate(8)
		market.Reserves = market.Reserves.Add(reserves).Truncate(8)
		market.TotalCash = market.TotalCash.Add(repayAmount)

		// keep the flywheel moving
		e = w.marketService.KeppFlywheelMoving(ctx, tx, market, snapshot.CreatedAt)
		if e != nil {
			return e
		}

		//update interest index
		e = w.borrowService.UpdateMarketInterestIndex(ctx, tx, market, market.BlockNumber)
		if e != nil {
			return e
		}

		return nil
	})
}
