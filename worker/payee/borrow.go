package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// handle borrow event
func (w *Payee) handleBorrowEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("event", "borrow")
	ctx = logger.WithContext(ctx, log)

	var (
		assetID      string
		borrowAmount decimal.Decimal
	)
	{
		var asset uuid.UUID
		_, e := mtg.Scan(body, &asset, &borrowAmount)
		if err := compound.Require(e == nil, "payee/mtgscan"); err != nil {
			log.WithError(err).Infoln("skip: scan memo failed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeBorrow, core.ErrInvalidArgument)
		}

		borrowAmount = borrowAmount.Truncate(8)
		assetID = asset.String()
		log = logger.FromContext(ctx).WithFields(logrus.Fields{
			"asset_id": assetID,
			"amount":   borrowAmount,
		})
		ctx = logger.WithContext(ctx, log)
	}

	market, err := w.mustGetMarket(ctx, assetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketNotFound)
	}

	if market.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	// accrue interest
	AccrueInterest(ctx, market, output.CreatedAt)

	borrow, err := w.borrowStore.Find(ctx, userID, assetID)
	if err != nil {
		log.WithError(err).Errorln("borrows.Find")
		return err
	}

	if borrow.ID == 0 {
		//new borrow record
		borrow = &core.Borrow{
			UserID:        userID,
			AssetID:       market.AssetID,
			InterestIndex: market.BorrowIndex,
		}

		if err := w.borrowStore.Create(ctx, borrow); err != nil {
			log.WithError(err).Errorln("borrows.Create")
			return err
		}
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.Find")
		return err
	}

	if tx.ID == 0 {
		if err := compound.Require(!market.IsMarketClosed(), "payee/market-closed", compound.FlagRefund); err != nil {
			log.WithError(err).Infoln("market closed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeBorrow, core.ErrMarketClosed)
		}

		liquidity, err := w.accountService.CalculateAccountLiquidity(ctx, userID, market)
		if err != nil {
			log.WithError(err).Errorln("CalculateAccountLiquidity")
			return err
		}

		if err := compound.Require(
			market.BorrowAllowed(borrowAmount) && borrowAmount.Mul(market.Price).LessThanOrEqual(liquidity),
			"payee/borrow-denied",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Infoln("borrow not allowed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeBorrow, core.ErrBorrowNotAllowed)
		}

		extra := core.NewTransactionExtra()
		extra.Put("asset_id", assetID)
		extra.Put("amount", borrowAmount)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeBorrow, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	//update borrow account
	if output.ID > borrow.Version {
		borrow.Principal = compound.BorrowBalance(ctx, borrow, market).Add(borrowAmount)
		borrow.InterestIndex = market.BorrowIndex
		if err := w.borrowStore.Update(ctx, borrow, output.ID); err != nil {
			log.WithError(err).Errorln("borrows.Update")
			return err
		}
	}

	//transfer borrowed asset
	if err := w.transferOut(
		ctx,
		userID,
		followID,
		output.TraceID,
		assetID,
		borrowAmount,
		&core.TransferAction{
			Source:   core.ActionTypeBorrowTransfer,
			FollowID: followID,
		},
	); err != nil {
		log.WithError(err).Errorln("transferOut")
		return err
	}

	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(borrowAmount).Truncate(compound.MaxPricision)
		market.TotalBorrows = market.TotalBorrows.Add(borrowAmount).Truncate(compound.MaxPricision)
		//update interest
		AccrueInterest(ctx, market, output.CreatedAt)
		// update market
		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	return nil
}
