package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from user, refund if error
var handleBorrowRepayEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	amount := snapshot.Amount.Abs()
	userID := snapshot.OpponentID

	market, e := w.marketStore.Find(ctx, snapshot.AssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//update borrow
	borrow, e := w.borrowStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	return w.db.Tx(func(tx *db.DB) error {
		borrow.Principal = borrow.Principal.Sub(amount)
		if borrow.Principal.LessThan(decimal.Zero) {
			borrow.Principal = decimal.Zero
		}
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			return e
		}

		//TODO 更新流动性

		return nil
	})
}
