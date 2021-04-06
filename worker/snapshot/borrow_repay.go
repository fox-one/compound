package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

// handle borrow repay event
func (w *Payee) handleRepayEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {

	log := logger.FromContext(ctx).WithField("worker", "borrow_repay")

	repayAmount := output.Amount
	assetID := output.AssetID

	log.Infoln(":asset:", output.AssetID, "amount:", repayAmount)
	market, isRecordNotFound, e := w.marketStore.Find(ctx, assetID)
	if isRecordNotFound {
		log.Warningln("market not found")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRepay, core.ErrMarketNotFound)
	}

	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRepay, core.ErrMarketClosed)
	}

	//update interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	borrow, isRecordNotFound, e := w.borrowStore.Find(ctx, userID, market.AssetID)
	if isRecordNotFound {
		log.Warningln("borrow not found")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRepay, core.ErrBorrowNotFound)
	}
	if e != nil {
		log.Errorln(e)
		return e
	}

	//update borrow info
	borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
	if e != nil {
		log.Errorln(e)
		return e
	}
	realRepaidBalance := repayAmount
	redundantAmount := repayAmount.Sub(borrowBalance).Truncate(8)
	newBalance := borrowBalance.Sub(repayAmount)
	newIndex := market.BorrowIndex
	if newBalance.LessThanOrEqual(decimal.Zero) {
		newBalance = decimal.Zero
		newIndex = decimal.Zero
		realRepaidBalance = borrowBalance
	}

	if output.ID > borrow.Version {
		borrow.Principal = newBalance.Truncate(16)
		borrow.InterestIndex = newIndex.Truncate(16)
		if e = w.borrowStore.Update(ctx, borrow, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	if output.ID > market.Version {
		market.TotalBorrows = market.TotalBorrows.Sub(realRepaidBalance).Truncate(16)
		market.TotalCash = market.TotalCash.Add(realRepaidBalance).Truncate(16)
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

	// add transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
		UserID:        borrow.UserID,
		AssetID:       borrow.AssetID,
		Principal:     borrow.Principal,
		InterestIndex: borrow.InterestIndex,
	})
	transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeRepay, output, extra)
	if e = w.transactionStore.Create(ctx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	if redundantAmount.GreaterThan(decimal.Zero) {
		refundAmount := redundantAmount
		transferAction := core.TransferAction{
			Source:   core.ActionTypeRepayRefundTransfer,
			FollowID: followID,
		}

		return w.transferOut(ctx, userID, followID, output.TraceID, assetID, refundAmount, &transferAction)
	}

	return nil
}
