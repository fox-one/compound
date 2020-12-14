package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"
	"strings"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

var handleBorrowEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "borrow_event")

	symbol := strings.ToUpper(action[core.ActionKeySymbol])
	userID := snapshot.OpponentID

	borrowAmount, e := decimal.NewFromString(action[core.ActionKeyAmount])
	if e != nil {
		log.Errorln("parse amount error:", e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		log.Errorln("query market error:", e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		return e
	}

	if !w.borrowService.BorrowAllowed(ctx, borrowAmount, userID, market, snapshot.CreatedAt) {
		log.Errorln("borrow not allowed")
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrBorrowNotAllowed)
	}

	return w.db.Tx(func(tx *db.DB) error {
		market.TotalCash = market.TotalCash.Sub(borrowAmount).Truncate(16)
		market.TotalBorrows = market.TotalBorrows.Add(borrowAmount).Truncate(16)
		// update market
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		//update interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		borrow, e := w.borrowStore.Find(ctx, userID, market.AssetID)
		if e != nil {
			if gorm.IsRecordNotFoundError(e) {
				//new
				borrow = &core.Borrow{
					UserID:        userID,
					AssetID:       market.AssetID,
					Principal:     borrowAmount,
					InterestIndex: market.BorrowIndex}

				if e = w.borrowStore.Save(ctx, tx, borrow); e != nil {
					log.Errorln(e)
					return e
				}
			} else {
				return e
			}
		} else {
			//update borrow account
			borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
			if e != nil {
				log.Errorln(e)
				return e
			}

			newBorrowBalance := borrowBalance.Add(borrowAmount)
			borrow.Principal = newBorrowBalance.Truncate(16)
			borrow.InterestIndex = market.BorrowIndex.Truncate(16)
			e = w.borrowStore.Update(ctx, tx, borrow)
			if e != nil {
				log.Errorln(e)
				return e
			}
		}

		//transfer to user
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceBorrowTransfer
		memoStr, e := memo.Format()
		if e != nil {
			log.Errorln("memo format error:", e)
			return e
		}
		trace := id.UUIDFromString(fmt.Sprintf("borrow:%s", snapshot.TraceID))
		transfer := core.Transfer{
			AssetID:    market.AssetID,
			OpponentID: userID,
			Amount:     borrowAmount,
			TraceID:    trace,
			Memo:       memoStr,
		}

		if e = w.transferStore.Create(ctx, tx, &transfer); e != nil {
			log.Errorln(e)
			return e
		}

		return nil
	})
}
