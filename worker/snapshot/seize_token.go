package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleSeizeTokenEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "seize_token")

	var seizedUser uuid.UUID
	var seizedAsset uuid.UUID
	if _, err := mtg.Scan(body, &seizedUser, &seizedAsset); err != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrInvalidArgument, "")
	}

	seizedUserID := seizedUser.String()
	seizedAssetID := seizedAsset.String()

	userPayAmount := output.Amount.Abs()
	userPayAssetID := output.AssetID

	// to seize
	supplyMarket, e := w.marketStore.Find(ctx, seizedAssetID)
	if e != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}

	supplyExchangeRate, e := w.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		log.Errorln(e)
		return e
	}

	// to repay
	borrowMarket, e := w.marketStore.Find(ctx, userPayAssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}

	//supply market accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, supplyMarket, output.UpdatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	//borrow market accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, borrowMarket, output.UpdatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	supply, e := w.supplyStore.Find(ctx, seizedUserID, supplyMarket.CTokenAssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrSupplyNotFound, "")
	}

	borrow, e := w.borrowStore.Find(ctx, seizedUserID, borrowMarket.AssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrBorrowNotFound, "")
	}

	borrowPrice, e := w.priceService.GetCurrentUnderlyingPrice(ctx, borrowMarket)
	if e != nil {
		log.Errorln(e)
		return e
	}

	if borrowPrice.LessThanOrEqual(decimal.Zero) {
		log.Errorln(e)
		return e
	}

	supplyPrice, e := w.priceService.GetCurrentUnderlyingPrice(ctx, supplyMarket)
	if e != nil {
		log.Errorln(e)
		return e
	}
	if supplyPrice.LessThanOrEqual(decimal.Zero) {
		log.Errorln(e)
		return e
	}

	// refund to liquidator if seize not allowed
	if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, output.UpdatedAt) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrSeizeNotAllowed, "")
	}

	return w.db.Tx(func(tx *db.DB) error {
		borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, borrowMarket)
		if e != nil {
			log.Errorln(e)
			return e
		}

		maxSeize := supply.Collaterals.Mul(supplyExchangeRate).Mul(supplyMarket.CloseFactor)
		seizedPrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
		maxSeizeValue := maxSeize.Mul(seizedPrice)
		repayValue := userPayAmount.Mul(borrowPrice)
		borrowBalanceValue := borrowBalance.Mul(borrowPrice)
		seizedAmount := repayValue.Div(seizedPrice)
		if repayValue.GreaterThan(maxSeizeValue) {
			repayValue = maxSeizeValue
			seizedAmount = repayValue.Div(seizedPrice)
		}

		if repayValue.GreaterThan(borrowBalanceValue) {
			repayValue = borrowBalanceValue
			seizedAmount = repayValue.Div(seizedPrice)
		}

		seizedCTokens := seizedAmount.Div(supplyExchangeRate)
		//update supply
		supply.Collaterals = supply.Collaterals.Sub(seizedCTokens).Truncate(16)
		if e = w.supplyStore.Update(ctx, tx, supply); e != nil {
			log.Errorln(e)
			return e
		}

		//update supply market ctokens
		supplyMarket.TotalCash = supplyMarket.TotalCash.Sub(seizedAmount).Truncate(16)
		supplyMarket.CTokens = supplyMarket.CTokens.Sub(seizedCTokens).Truncate(16)
		if e = w.marketStore.Update(ctx, tx, supplyMarket); e != nil {
			log.Errorln(e)
			return e
		}

		// update borrow account and borrow market
		reallyRepayAmount := repayValue.Div(borrowPrice)
		redundantAmount := userPayAmount.Sub(reallyRepayAmount)
		newBorrowBalance := borrowBalance.Sub(reallyRepayAmount).Truncate(8)
		newIndex := borrowMarket.BorrowIndex
		if newBorrowBalance.LessThanOrEqual(decimal.Zero) {
			newBorrowBalance = decimal.Zero
			newIndex = decimal.Zero
		}
		borrow.Principal = newBorrowBalance.Truncate(16)
		borrow.InterestIndex = newIndex.Truncate(16)
		if e = w.borrowStore.Update(ctx, tx, borrow); e != nil {
			log.Errorln(e)
			return e
		}

		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Sub(reallyRepayAmount).Truncate(16)
		borrowMarket.TotalCash = borrowMarket.TotalCash.Add(reallyRepayAmount).Truncate(16)
		if e = w.marketStore.Update(ctx, tx, borrowMarket); e != nil {
			log.Errorln(e)
			return e
		}

		//supply market accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, supplyMarket, output.UpdatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		//borrow market accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, borrowMarket, output.UpdatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		transferAction := core.TransferAction{
			Source:        core.ActionTypeSeizeTokenTransfer,
			TransactionID: followID,
		}

		if e = w.transferOut(ctx, userID, followID, output.TraceID, supplyMarket.AssetID, seizedAmount, &transferAction); e != nil {
			return e
		}

		//refund redundant assets to liquidator
		if redundantAmount.GreaterThan(decimal.Zero) {
			refundAmount := redundantAmount.Truncate(8)

			refundTransferAction := core.TransferAction{
				Source:        core.ActionTypeRefundTransfer,
				TransactionID: followID,
			}

			if e = w.transferOut(ctx, userID, followID, output.TraceID, output.AssetID, refundAmount, &refundTransferAction); e != nil {
				return e
			}
		}

		return nil
	})
}
