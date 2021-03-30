package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// handle borrow event
func (w *Payee) handleBorrowEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {

	log := logger.FromContext(ctx).WithField("worker", "borrow")

	var asset uuid.UUID
	var borrowAmount decimal.Decimal
	if _, err := mtg.Scan(body, &asset, &borrowAmount); err != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeBorrow, core.ErrInvalidArgument)
	}

	assetID := asset.String()
	log.Infoln("borrow, asset:", assetID, ":amount:", borrowAmount)
	market, isRecordNotFound, e := w.marketStore.Find(ctx, assetID)
	if isRecordNotFound {
		log.Warningln("market not found, refund")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketNotFound)
	}

	if e != nil {
		log.Errorln("query market error:", e)
		return e
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketClosed)
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
	}

	if !w.borrowService.BorrowAllowed(ctx, borrowAmount, userID, market, output.CreatedAt) {
		log.Errorln("borrow not allowed")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeBorrow, core.ErrBorrowNotAllowed)
	}

	market.TotalCash = market.TotalCash.Sub(borrowAmount).Truncate(16)
	market.TotalBorrows = market.TotalBorrows.Add(borrowAmount).Truncate(16)
	// update market
	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	borrow, isRecordNotFound, e := w.borrowStore.Find(ctx, userID, market.AssetID)
	if e != nil {
		if isRecordNotFound {
			//new borrow record
			borrow = &core.Borrow{
				UserID:        userID,
				AssetID:       market.AssetID,
				Principal:     borrowAmount,
				InterestIndex: market.BorrowIndex}

			if e = w.borrowStore.Save(ctx, borrow); e != nil {
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
		e = w.borrowStore.Update(ctx, borrow, output.ID)
		if e != nil {
			log.Errorln(e)
			return e
		}
	}

	//transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyAssetID, assetID)
	extra.Put(core.TransactionKeyAmount, borrowAmount)
	transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeBorrow, output, &extra)
	if e = w.transactionStore.Create(ctx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	//transfer borrowed asset
	transferAction := core.TransferAction{
		Source:   core.ActionTypeBorrowTransfer,
		FollowID: followID,
	}
	return w.transferOut(ctx, userID, followID, output.TraceID, assetID, borrowAmount, &transferAction)
}
