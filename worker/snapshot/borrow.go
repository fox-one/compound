package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleBorrowEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "borrow")

	var asset uuid.UUID
	var borrowAmount decimal.Decimal
	if _, err := mtg.Scan(body, &asset, &borrowAmount); err != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrInvalidArgument, "")
	}

	assetID := asset.String()
	market, e := w.marketStore.Find(ctx, assetID)
	if e != nil {
		log.Errorln("query market error:", e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, output.UpdatedAt); e != nil {
		return e
	}

	if !w.borrowService.BorrowAllowed(ctx, borrowAmount, userID, market, output.UpdatedAt) {
		log.Errorln("borrow not allowed")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrBorrowNotAllowed, "")
	}

	return w.db.Tx(func(tx *db.DB) error {
		market.TotalCash = market.TotalCash.Sub(borrowAmount)
		market.TotalBorrows = market.TotalBorrows.Add(borrowAmount)
		// update market
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		//update interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, output.UpdatedAt); e != nil {
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
			borrow.Principal = newBorrowBalance
			borrow.InterestIndex = market.BorrowIndex
			e = w.borrowStore.Update(ctx, tx, borrow)
			if e != nil {
				log.Errorln(e)
				return e
			}
		}

		transferAction := core.TransferAction{
			Source:        core.ActionTypeBorrowTransfer,
			TransactionID: followID,
		}

		return w.transferOut(ctx, userID, followID, output.TraceID, assetID, borrowAmount, &transferAction)
	})
}
