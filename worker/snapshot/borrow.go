package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
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

	borrowAmount = borrowAmount.Truncate(8)
	assetID := asset.String()
	log.Infoln("borrow, asset:", assetID, ":amount:", borrowAmount)

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		market, e := w.marketStore.Find(ctx, assetID)
		if e != nil {
			log.WithError(e).Errorln("find market error")
			return e
		}

		borrow, e := w.borrowStore.Find(ctx, userID, market.AssetID)
		if e != nil {
			log.Errorln(e)
			return e
		}

		cs := core.NewContextSnapshot(nil, borrow, nil, market)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeBorrow, output, cs)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	contextSnapshot, e := tx.UnmarshalContextSnapshot()
	if e != nil {
		return e
	}

	market := contextSnapshot.BorrowMarket
	if market == nil || market.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketNotFound)
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketClosed)
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
	}

	if !w.borrowService.BorrowAllowed(ctx, borrowAmount, userID, market, output.CreatedAt) {
		log.Errorln("borrow not allowed")
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeBorrow, core.ErrBorrowNotAllowed)
	}

	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(borrowAmount).Truncate(16)
		market.TotalBorrows = market.TotalBorrows.Add(borrowAmount).Truncate(16)
		// update market
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	borrow := contextSnapshot.Borrow
	if borrow == nil || borrow.ID == 0 {
		//new borrow record
		borrow = &core.Borrow{
			UserID:        userID,
			AssetID:       market.AssetID,
			Principal:     borrowAmount,
			InterestIndex: market.BorrowIndex,
			Version:       output.ID}

		if e = w.borrowStore.Save(ctx, borrow); e != nil {
			log.Errorln(e)
			return e
		}
	} else {
		//update borrow account
		if output.ID > borrow.Version {
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
	}

	//borrow transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyAssetID, assetID)
	extra.Put(core.TransactionKeyAmount, borrowAmount)
	extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
		UserID:        borrow.UserID,
		AssetID:       borrow.AssetID,
		Principal:     borrow.Principal,
		InterestIndex: borrow.InterestIndex,
	})
	tx.SetExtraData(extra)
	tx.Status = core.TransactionStatusComplete
	if e = w.transactionStore.Update(ctx, tx); e != nil {
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
