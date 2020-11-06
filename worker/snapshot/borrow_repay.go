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
	amount := snapshot.Amount.Abs()
	userID := snapshot.OpponentID

	market, e := w.marketStore.Find(ctx, snapshot.AssetID, "")
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	//update borrow
	borrow, e := w.borrowStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot)
	}

	return w.db.Tx(func(tx *db.DB) error {
		interestChanged := amount.Mul(borrow.InterestBalance.Div(borrow.Principal))

		borrow.Principal = borrow.Principal.Sub(amount)
		borrow.InterestBalance = borrow.InterestBalance.Sub(interestChanged)
		if borrow.Principal.LessThan(decimal.Zero) {
			borrow.Principal = decimal.Zero
			borrow.InterestBalance = decimal.Zero
		}
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			return e
		}

		//计提保留金,转账到block钱包
		reserve := interestChanged.Mul(market.ReserveFactor)
		trace := id.UUIDFromString(fmt.Sprintf("reserve:%s", snapshot.TraceID))
		input := mixin.TransferInput{
			AssetID:    market.AssetID,
			OpponentID: w.blockWallet.Client.ClientID,
			Amount:     reserve,
			TraceID:    trace,
		}

		if !w.walletService.VerifyPayment(ctx, &input) {
			memo := core.NewAction()
			memo[core.ActionKeyService] = core.ActionServiceReserve
			memoStr, e := memo.Format()
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
	})
}
