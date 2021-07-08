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

	borrowAmount = borrowAmount.Truncate(8)
	assetID := asset.String()
	log.Infoln("borrow, asset:", assetID, ":amount:", borrowAmount)

	market, e := w.marketStore.Find(ctx, assetID)
	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketNotFound)
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, assetID)
	if e != nil {
		return e
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		if w.marketService.IsMarketClosed(ctx, market) {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketClosed)
		}

		if !w.borrowService.BorrowAllowed(ctx, borrowAmount, userID, market, output.CreatedAt) {
			log.Errorln("borrow not allowed")
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeBorrow, core.ErrBorrowNotAllowed)
		}

		newBorrowBalance := decimal.Zero
		if borrow.ID == 0 {
			newBorrowBalance = borrowAmount.Truncate(16)
		} else {
			borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
			if e != nil {
				log.Errorln(e)
				return e
			}
			newBorrowBalance = borrowBalance.Add(borrowAmount).Truncate(16)
		}

		extra := core.NewTransactionExtra()
		extra.Put(core.TransactionKeyAssetID, assetID)
		extra.Put(core.TransactionKeyAmount, borrowAmount)
		extra.Put("new_borrow_balance", newBorrowBalance)
		extra.Put("new_borrow_index", market.BorrowIndex)
		extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
			UserID:        userID,
			AssetID:       assetID,
			Principal:     newBorrowBalance,
			InterestIndex: market.BorrowIndex,
		})

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeBorrow, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		NewBorrowBalance decimal.Decimal `json:"new_borrow_balance"`
		NewBorrowIndex   decimal.Decimal `json:"new_borrow_index"`
	}

	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	if borrow.ID == 0 {
		//new borrow record
		borrow = &core.Borrow{
			UserID:        userID,
			AssetID:       market.AssetID,
			Principal:     extra.NewBorrowBalance,
			InterestIndex: extra.NewBorrowIndex,
			Version:       output.ID}

		if e = w.borrowStore.Save(ctx, borrow); e != nil {
			log.Errorln(e)
			return e
		}
	} else {
		//update borrow account
		if output.ID > borrow.Version {
			borrow.Principal = extra.NewBorrowBalance
			borrow.InterestIndex = extra.NewBorrowIndex
			e = w.borrowStore.Update(ctx, borrow, output.ID)
			if e != nil {
				log.Errorln(e)
				return e
			}
		}
	}

	//transfer borrowed asset
	transferAction := core.TransferAction{
		Source:   core.ActionTypeBorrowTransfer,
		FollowID: followID,
	}
	if err := w.transferOut(ctx, userID, followID, output.TraceID, assetID, borrowAmount, &transferAction); err != nil {
		return err
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

	return nil
}
