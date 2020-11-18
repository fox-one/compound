package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"
	"strings"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

var handleBorrowEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "borrow_event")

	symbol := strings.ToUpper(action[core.ActionKeySymbol])
	userID := action[core.ActionKeyUser]
	amount, e := decimal.NewFromString(action[core.ActionKeyAmount])
	if e != nil {
		log.Errorln("parse amount error:", e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}

	market, e := w.marketStore.FindByCToken(ctx, symbol)
	if e != nil {
		log.Errorln("query market error:", e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		return e
	}

	if !w.borrowService.BorrowAllowed(ctx, amount, userID, market) {
		log.Errorln("borrow not allowed")
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrBorrowNotAllowed)
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
			return e
		}

		input.Memo = memoStr
		_, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin)
		if e != nil {
			log.Errorln("transfer borrow asset to user error:", e)
			return e
		}
	}

	return nil
}

var handleBorrowTransferEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return w.db.Tx(func(tx *db.DB) error {
		userID := snapshot.OpponentID
		borrowAmount := snapshot.Amount.Abs()

		market, e := w.marketStore.Find(ctx, snapshot.AssetID)
		if e != nil {
			return e
		}

		//update interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			return e
		}

		market.TotalCash = market.TotalCash.Sub(borrowAmount)
		market.TotalBorrows = market.TotalBorrows.Add(borrowAmount)
		// update market
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			return e
		}

		borrow, e := w.borrowStore.Find(ctx, userID, market.Symbol)
		if e != nil {
			//new
			borrow := core.Borrow{
				UserID:        userID,
				Symbol:        market.Symbol,
				Principal:     borrowAmount,
				InterestIndex: market.BorrowIndex}

			if e = w.borrowStore.Save(ctx, tx, &borrow); e != nil {
				return e
			}

			return nil
		}

		//update borrow account
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
		if e != nil {
			return e
		}

		newBorrowBalance := borrowBalance.Add(borrowAmount)
		borrow.Principal = newBorrowBalance
		borrow.InterestIndex = market.BorrowIndex
		e = w.borrowStore.Update(ctx, tx, borrow)
		if e != nil {
			return e
		}

		return nil
	})
}
