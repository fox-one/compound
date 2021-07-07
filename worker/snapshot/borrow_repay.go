package snapshot

import (
	"compound/core"
	"context"
	"encoding/json"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// handle borrow repay event
func (w *Payee) handleRepayEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"worker": "borrow_repay",
		"asset":  output.AssetID,
		"amount": output.Amount,
	})

	log.Debugln("repaying")

	market, e := w.marketStore.Find(ctx, output.AssetID)
	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if marekt.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeRepay, core.ErrMarketNotFound)
	}

	//update interest
	if e := w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.WithError(e).Errorln("AccrueInterest")
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, market.AssetID)
	if e != nil {
		log.WithError(e).Errorln("find borrow error")
		return e
	}
	if borrow.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeRepay, core.ErrMarketNotFound)
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		cs := core.NewContextSnapshot(nil, borrow, nil, market)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeRepay, output, cs)

		if w.marketService.IsMarketClosed(ctx, market) {
			tx.Status = core.TransactionStatusAbort
		} else {
			tx.Status = core.TransactionStatusComplete

			//update borrow info
			borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
			if e != nil {
				log.WithError(e).Errorln("BorrowBalance")
				return e
			}

			repayAmount := extra.Amount
			if borrowBalance.LessThan(extra.Amount) {
				repayAmount = borrowBalance
			}
			extra := core.NewTransactionExtra()
			extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
				UserID:  userID,
				AssetID: borrow.AssetID,
				Amount:  repayAmount,
			})
			tx.SetExtraData(extra)
		}

		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	if tx.Status == core.TransactionStatusAbort {
		return w.handleRefundEvent(ctx, output, userID, followID, action, errCode)
	}

	var extra struct {
		Borrow struct {
			Amount decimal.Decimal `json:"amount"`
		} `json:"borrow"`
	}

	if err := json.Unmarshal(tx.Data, &extra); err != nil {
		panic(err)
	}

	if refundAmount := output.Amount.Sub(extra.Borrow.Amount).Truncate(8); refundAmount.IsPositive() {
		transferAction := core.TransferAction{
			Source:   core.ActionTypeRepayRefundTransfer,
			FollowID: followID,
		}

		return w.transferOut(ctx, userID, followID, output.TraceID, output.AssetID, refundAmount, &transferAction)
	}

	if output.ID > borrow.Version {
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
		if e != nil {
			log.WithError(e).Errorln("BorrowBalance")
			return e
		}
		borrow.Principal = borrowBalance.Sub(extra.Borrow.Amount)
		borrow.InterestIndex = market.BorrowIndex
		if borrow.Principal.IsZero() {
			borrow.InterestIndex = decimal.Zero
		}
		if e := w.borrowStore.Update(ctx, borrow, output.ID); e != nil {
			log.WithError(e).Errorln("borrow.Update")
			return e
		}
	}

	if output.ID > market.Version {
		market.TotalBorrows = market.TotalBorrows.Sub(extra.Borrow.Amount)
		market.TotalCash = market.TotalCash.Add(extra.Borrow.Amount)

		// market transaction
		marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
		if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
			log.WithError(e).Errorln("create transaction error")
			return e
		}

		if e := w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.WithError(e).Errorln("market.Update")
			return e
		}
	}

	return nil
}
