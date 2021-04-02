package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// handle liquidation event
func (w *Payee) handleLiquidationEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "seize_token")

	liquidator := userID
	var seizedAddress uuid.UUID
	var seizedCTokenAsset uuid.UUID
	if _, err := mtg.Scan(body, &seizedAddress, &seizedCTokenAsset); err != nil {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrInvalidArgument)
	}

	// check market close status
	if w.marketService.HasClosedMarkets(ctx) {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrMarketClosed)
	}

	seizedUser, e := w.userStore.FindByAddress(ctx, seizedAddress.String())
	if e != nil {
		if gorm.IsRecordNotFoundError(e) {
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrInvalidArgument)
		}
		return e
	}

	// check allowlist
	needAllowListCheck, e := w.allowListService.IsScopeInAllowList(ctx, core.OSLiquidation)
	if e != nil {
		return e
	}
	if needAllowListCheck {
		userAllowed, e := w.allowListService.CheckAllowList(ctx, seizedUser.UserID, core.OSLiquidation)
		if e != nil {
			return e
		}
		if !userAllowed {
			// not allowed, refund
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrOperationForbidden)
		}
	}

	seizedUserID := seizedUser.UserID
	seizedCTokenAssetID := seizedCTokenAsset.String()

	userPayAmount := output.Amount
	userPayAssetID := output.AssetID

	log.Infof("seizedUser:%s, seizedAsset:%s, payAsset:%s, payAmount:%s", seizedUserID, seizedCTokenAssetID, userPayAssetID, userPayAmount)

	// supply market
	supplyMarket, isRecordNotFound, e := w.marketStore.FindByCToken(ctx, seizedCTokenAssetID)
	if isRecordNotFound {
		log.Warningln("supply market not found")
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}
	if e != nil {
		log.WithError(e).Errorln("find supply market error")
		return e
	}

	// borrow market
	borrowMarket, isRecordNotFound, e := w.marketStore.Find(ctx, userPayAssetID)
	if isRecordNotFound {
		log.Warningln("borrow market not found")
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}
	if e != nil {
		log.WithError(e).Errorln("find borrow market error")
		return e
	}

	//supply market accrue interest
	if e = w.marketService.AccrueInterest(ctx, supplyMarket, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	//borrow market accrue interest
	if e = w.marketService.AccrueInterest(ctx, borrowMarket, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	supplyExchangeRate, e := w.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		log.Errorln(e)
		return e
	}

	// supply
	supply, isRecordNotFound, e := w.supplyStore.Find(ctx, seizedUserID, supplyMarket.CTokenAssetID)
	if isRecordNotFound {
		log.Warningln("supply not found")
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrSupplyNotFound)
	}

	if e != nil {
		log.WithError(e).Errorln("find supply error")
		return e
	}

	// borrow
	borrow, isRecordNotFound, e := w.borrowStore.Find(ctx, seizedUserID, borrowMarket.AssetID)
	if isRecordNotFound {
		log.Warningln("borrow not found")
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrBorrowNotFound)
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
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrSeizeNotAllowed)
	}

	borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, borrowMarket)
	if e != nil {
		log.Errorln(e)
		return e
	}

	// calculate values
	//ctokenValue = ctokenAmount / exchange_rate * price
	maxSeize := supply.Collaterals.Mul(supplyExchangeRate).Mul(supplyMarket.CloseFactor).Truncate(16)
	seizedPrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive)).Truncate(16)
	maxSeizeValue := maxSeize.Mul(seizedPrice).Truncate(16)
	repayValue := userPayAmount.Mul(borrowPrice).Truncate(16)
	borrowBalanceValue := borrowBalance.Mul(borrowPrice).Truncate(16)
	seizedAmount := repayValue.Div(seizedPrice).Truncate(16)
	if repayValue.GreaterThan(maxSeizeValue) {
		repayValue = maxSeizeValue
		seizedAmount = repayValue.Div(seizedPrice)
	}

	if repayValue.GreaterThan(borrowBalanceValue) {
		repayValue = borrowBalanceValue
		seizedAmount = repayValue.Div(seizedPrice)
	}

	seizedCTokens := seizedAmount.Div(supplyExchangeRate).Truncate(8)
	//update supply
	supply.Collaterals = supply.Collaterals.Sub(seizedCTokens).Truncate(16)
	if e = w.supplyStore.Update(ctx, supply, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	//update supply market ctokens
	if e = w.marketStore.Update(ctx, supplyMarket, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	// update borrow account
	reallyRepayAmount := repayValue.Div(borrowPrice).Truncate(16)
	redundantAmount := userPayAmount.Sub(reallyRepayAmount).Truncate(8)
	newBorrowBalance := borrowBalance.Sub(reallyRepayAmount).Truncate(16)
	newIndex := borrowMarket.BorrowIndex
	if newBorrowBalance.LessThanOrEqual(decimal.Zero) {
		newBorrowBalance = decimal.Zero
		newIndex = decimal.Zero
	}
	borrow.Principal = newBorrowBalance.Truncate(16)
	borrow.InterestIndex = newIndex.Truncate(16)
	if e = w.borrowStore.Update(ctx, borrow, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	// update borrow market
	borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Sub(reallyRepayAmount).Truncate(16)
	borrowMarket.TotalCash = borrowMarket.TotalCash.Add(reallyRepayAmount).Truncate(16)
	if e = w.marketStore.Update(ctx, borrowMarket, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	// add transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyCTokenAssetID, seizedCTokenAssetID)
	extra.Put(core.TransactionKeyAmount, seizedCTokens)
	extra.Put(core.TransactionKeyPrice, seizedPrice)
	if redundantAmount.GreaterThan(decimal.Zero) {
		extra.Put(core.TransactionKeyRefund, redundantAmount)
	} else {
		extra.Put(core.TransactionKeyRefund, decimal.Zero)
	}

	transaction := core.BuildTransactionFromOutput(ctx, liquidator, followID, core.ActionTypeLiquidate, output, &extra)
	if e = w.transactionStore.Create(ctx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// transfer seized ctoken to liquidator
	transferAction := core.TransferAction{
		Source:   core.ActionTypeLiquidateTransfer,
		FollowID: followID,
	}
	if e = w.transferOut(ctx, liquidator, followID, output.TraceID, supplyMarket.CTokenAssetID, seizedCTokens, &transferAction); e != nil {
		return e
	}

	//refund redundant assets to liquidator
	if redundantAmount.GreaterThan(decimal.Zero) {
		refundAmount := redundantAmount

		refundTransferAction := core.TransferAction{
			Source:   core.ActionTypeLiquidateRefundTransfer,
			FollowID: followID,
		}
		if e = w.transferOut(ctx, liquidator, followID, output.TraceID, output.AssetID, refundAmount, &refundTransferAction); e != nil {
			return e
		}
	}

	return nil
}
