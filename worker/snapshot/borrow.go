package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from system, ignore if error
var handleBorrowEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "borrow_event")

	market, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID, "")
	if e != nil {
		log.Errorln("query market error:", e)
		return nil
	}

	// symbol := action[core.ActionKeySymbol]
	userID := action[core.ActionKeyUser]
	amount, err := decimal.NewFromString(action[core.ActionKeyAmount])
	if err != nil {
		log.Errorln("parse amount error:", e)
		return nil
	}

	if !w.borrowService.BorrowAllowed(ctx, amount, userID, market) {
		log.Errorln("borrow not allowed")
		return nil
	}

	//transfer borrow asset to user
	trace := id.UUIDFromString(fmt.Sprintf("borrow:%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    market.AssetID,
		OpponentID: userID,
		Amount:     amount,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceBorrowTransfer
		memoStr, e := memo.Format()
		if e != nil {
			log.Errorln("memo format error:", e)
			return nil
		}

		input.Memo = memoStr
		_, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin)
		if e != nil {
			log.Errorln("transfer borrow asset to user error:", e)
			return nil
		}
	}

	return nil
}

var handleBorrowTransferEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := snapshot.OpponentID
	borrowAmount := snapshot.Amount.Abs()

	market, e := w.marketStore.Find(ctx, snapshot.AssetID, "")
	if e != nil {
		return e
	}

	return w.db.Tx(func(tx *db.DB) error {
		//update or insert borrow
		borrow, e := w.borrowStore.Find(ctx, userID, market.Symbol)
		if e != nil {
			//insert new
			borrow := core.Borrow{
				UserID:    userID,
				Symbol:    market.Symbol,
				Principal: borrowAmount,
			}

			e = w.borrowStore.Save(ctx, tx, &borrow)
			if e != nil {
				return e
			}
		} else {
			//update
			borrow.Principal = borrow.Principal.Add(borrowAmount)
			e = w.borrowStore.Update(ctx, tx, borrow)
			if e != nil {
				return e
			}
		}
		return nil
	})
}
