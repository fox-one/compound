package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"
	"encoding/json"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// handle borrow repay event
func (w *Payee) handleRepayEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("event", "borrow_repay")
	ctx = logger.WithContext(ctx, log)

	market, err := w.mustGetMarket(ctx, output.AssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeRepay, core.ErrMarketNotFound)
	}

	if market.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	//update interest
	AccrueInterest(ctx, market, output.CreatedAt)

	borrow, err := w.mustGetBorrow(ctx, userID, output.AssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeRepay, core.ErrBorrowNotFound)
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.Find")
		return err
	}

	if tx.ID == 0 {
		borrowBalance := compound.BorrowBalance(ctx, borrow, market)
		repayAmount := output.Amount
		if repayAmount.GreaterThan(borrowBalance) {
			repayAmount = borrowBalance
		}

		extra := core.NewTransactionExtra()
		extra.Put("repay_amount", repayAmount)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeRepay, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			log.WithError(err).Errorln("transactions.Create")
			return err
		}
	}

	var extra struct {
		RepayAmount decimal.Decimal `json:"repay_amount"`
	}
	if err := json.Unmarshal(tx.Data, &extra); err != nil {
		log.WithError(err).Errorln("Unmarshal extra")
		return err
	}

	if refundAmount := output.Amount.Sub(extra.RepayAmount).Truncate(8); refundAmount.IsPositive() {
		if err := w.transferOut(
			ctx,
			userID,
			followID,
			output.TraceID,
			output.AssetID,
			refundAmount,
			&core.TransferAction{
				Source:   core.ActionTypeRepayRefundTransfer,
				FollowID: followID,
			},
		); err != nil {
			return err
		}
	}

	if output.ID > borrow.Version {
		borrow.Principal = compound.BorrowBalance(ctx, borrow, market).Sub(extra.RepayAmount)
		borrow.InterestIndex = market.BorrowIndex
		if err := w.borrowStore.Update(ctx, borrow, output.ID); err != nil {
			log.WithError(err).Errorln("borrows.Update")
			return err
		}
	}

	if output.ID > market.Version {
		market.TotalBorrows = market.TotalBorrows.Sub(extra.RepayAmount).Truncate(compound.MaxPricision)
		market.TotalCash = market.TotalCash.Add(extra.RepayAmount).Truncate(compound.MaxPricision)
		if w.sysversion > 0 && market.TotalBorrows.IsNegative() {
			market.TotalBorrows = decimal.Zero
		}

		//update interest
		AccrueInterest(ctx, market, output.CreatedAt)
		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	return nil
}
