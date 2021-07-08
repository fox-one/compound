package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// handle borrow repay event
func (w *Payee) handleRepayEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {

	log := logger.FromContext(ctx).WithField("worker", "borrow_repay")

	repayAmount := output.Amount
	assetID := output.AssetID

	log.Infoln(":asset:", output.AssetID, "amount:", repayAmount)

	market, e := w.marketStore.Find(ctx, assetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRepay, core.ErrMarketNotFound)
	}

	//update interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, market.AssetID)
	if e != nil {
		return e
	}

	if borrow.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRepay, core.ErrBorrowNotFound)
	}

	transaction, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if transaction.ID == 0 {
		if w.marketService.IsMarketClosed(ctx, market) {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRepay, core.ErrMarketClosed)
		}

		//update borrow info
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
		if e != nil {
			log.Errorln(e)
			return e
		}

		newBalance := borrowBalance.Sub(repayAmount)
		newIndex := market.BorrowIndex
		if newBalance.LessThanOrEqual(decimal.Zero) {
			newBalance = decimal.Zero
			newIndex = decimal.Zero
			repayAmount = borrowBalance
		}

		extra := core.NewTransactionExtra()
		extra.Put("repay_amount", repayAmount.Truncate(16))
		extra.Put("new_balance", newBalance.Truncate(16))
		extra.Put("new_index", newIndex.Truncate(16))
		extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
			UserID:        borrow.UserID,
			AssetID:       borrow.AssetID,
			Principal:     newBalance,
			InterestIndex: newIndex,
		})

		transaction = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeRepay, output, extra)
		if e := w.transactionStore.Create(ctx, transaction); e != nil {
			return e
		}
	}

	var extra struct {
		RepayAmount decimal.Decimal `json:"repay_amount"`
		NewBalance  decimal.Decimal `json:"new_balance"`
		NewIndex    decimal.Decimal `json:"new_index"`
	}

	if e := transaction.UnmarshalExtraData(&extra); e != nil {
		return e
	}

	if output.ID > borrow.Version {
		borrow.Principal = extra.NewBalance
		borrow.InterestIndex = extra.NewIndex
		if e = w.borrowStore.Update(ctx, borrow, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	if refundAmount := output.Amount.Sub(extra.RepayAmount); refundAmount.GreaterThan(decimal.Zero) {
		transferAction := core.TransferAction{
			Source:   core.ActionTypeRepayRefundTransfer,
			FollowID: followID,
		}

		if e := w.transferOut(ctx, userID, followID, output.TraceID, assetID, refundAmount, &transferAction); e != nil {
			return e
		}
	}

	if output.ID > market.Version {
		market.TotalBorrows = market.TotalBorrows.Sub(extra.NewBalance).Truncate(16)
		market.TotalCash = market.TotalCash.Add(extra.NewBalance).Truncate(16)
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	return nil
}
