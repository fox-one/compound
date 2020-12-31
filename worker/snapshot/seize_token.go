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
	return w.db.Tx(func(tx *db.DB) error {
		log := logger.FromContext(ctx).WithField("worker", "seize_token")

		liquidator := userID
		var seizedAddress uuid.UUID
		var seizedAsset uuid.UUID
		if _, err := mtg.Scan(body, &seizedAddress, &seizedAsset); err != nil {
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ErrInvalidArgument, "")
		}

		seizedUser, e := w.userStore.FindByAddress(ctx, seizedAddress.String())
		if e != nil {
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ErrInvalidArgument, "")
		}

		seizedUserID := seizedUser.UserID
		seizedAssetID := seizedAsset.String()

		userPayAmount := output.Amount.Abs()
		userPayAssetID := output.AssetID

		log.Infof("seizedUser:%s, seizedAsset:%s, payAsset:%s, payAmount:%s", seizedUserID, seizedAssetID, userPayAssetID, userPayAmount)

		// to seize
		supplyMarket, isRecordNotFound, e := w.marketStore.Find(ctx, seizedAssetID)
		if isRecordNotFound {
			log.Warningln("supply market not found")
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ErrMarketNotFound, "")
		}
		if e != nil {
			log.WithError(e).Errorln("find supply market error")
			return e
		}

		supplyExchangeRate, e := w.marketService.CurExchangeRate(ctx, supplyMarket)
		if e != nil {
			log.Errorln(e)
			return e
		}

		// to repay
		borrowMarket, isRecordNotFound, e := w.marketStore.Find(ctx, userPayAssetID)
		if isRecordNotFound {
			log.Warningln("borrow market not found")
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ErrMarketNotFound, "")
		}
		if e != nil {
			log.WithError(e).Errorln("find borrow market error")
			return e
		}

		//supply market accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, supplyMarket, output.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		//borrow market accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, borrowMarket, output.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		supply, isRecordNotFound, e := w.supplyStore.Find(ctx, seizedUserID, supplyMarket.CTokenAssetID)
		if isRecordNotFound {
			log.Warningln("supply not found")
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ErrSupplyNotFound, "")
		}

		if e != nil {
			log.WithError(e).Errorln("find supply error")
			return e
		}

		borrow, isRecordNotFound, e := w.borrowStore.Find(ctx, seizedUserID, borrowMarket.AssetID)
		if isRecordNotFound {
			log.Warningln("borrow not found")
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ErrBorrowNotFound, "")
		}
		if e != nil {
			log.WithError(e).Errorln("find borrow error")
			return e
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
		if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, output.CreatedAt) {
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ErrSeizeNotAllowed, "")
		}

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

		seizedAmount = seizedAmount.Truncate(8)
		seizedCTokens := seizedAmount.Div(supplyExchangeRate).Truncate(16)
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
		reallyRepayAmount := repayValue.Div(borrowPrice).Truncate(16)
		redundantAmount := userPayAmount.Sub(reallyRepayAmount).Truncate(16)
		newBorrowBalance := borrowBalance.Sub(reallyRepayAmount).Truncate(16)
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
		if e = w.marketService.AccrueInterest(ctx, tx, supplyMarket, output.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		//borrow market accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, borrowMarket, output.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		// add transaction
		extra := core.NewTransactionExtra()
		extra.Put(core.TransactionKeyAssetID, seizedAsset)
		extra.Put(core.TransactionKeyAmount, seizedAmount)
		extra.Put(core.TransactionKeyPrice, seizedPrice)
		if redundantAmount.GreaterThan(decimal.Zero) {
			extra.Put(core.TransactionKeyRefund, redundantAmount.Truncate(8))
		} else {
			extra.Put(core.TransactionKeyRefund, decimal.Zero)
		}

		transaction := core.BuildTransactionFromOutput(ctx, liquidator, followID, core.ActionTypeSeizeToken, output, &extra)
		if e = w.transactionStore.Create(ctx, tx, transaction); e != nil {
			log.WithError(e).Errorln("create transaction error")
			return e
		}

		// transfer
		transferAction := core.TransferAction{
			Source:   core.ActionTypeSeizeTokenTransfer,
			FollowID: followID,
		}
		if e = w.transferOut(ctx, liquidator, followID, output.TraceID, supplyMarket.AssetID, seizedAmount, &transferAction); e != nil {
			return e
		}

		//refund redundant assets to liquidator
		if redundantAmount.GreaterThan(decimal.Zero) {
			refundAmount := redundantAmount.Truncate(8)

			refundTransferAction := core.TransferAction{
				Source:   core.ActionTypeSeizeRefundTransfer,
				FollowID: followID,
			}
			if e = w.transferOut(ctx, liquidator, followID, output.TraceID, output.AssetID, refundAmount, &refundTransferAction); e != nil {
				return e
			}
		}

		return nil
	})
}
