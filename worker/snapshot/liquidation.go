package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// handle liquidation event
func (w *Payee) handleLiquidationEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "liquidation")

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
		userAllowed, e := w.allowListService.CheckAllowList(ctx, userID, core.OSLiquidation)
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

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		supplyMarket, e := w.marketStore.FindByCToken(ctx, seizedCTokenAssetID)
		if e != nil {
			return e
		}

		borrowMarket, e := w.marketStore.Find(ctx, userPayAssetID)
		if e != nil {
			return e
		}

		supply, e := w.supplyStore.Find(ctx, seizedUserID, supplyMarket.CTokenAssetID)
		if e != nil {
			return e
		}

		borrow, e := w.borrowStore.Find(ctx, seizedUserID, borrowMarket.AssetID)
		if e != nil {
			return e
		}

		cs := core.NewContextSnapshot(supply, borrow, supplyMarket, borrowMarket)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeLiquidate, output, cs)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	contextSnapshot, e := tx.UnmarshalContextSnapshot()
	if e != nil {
		return e
	}

	supplyMarket := contextSnapshot.SupplyMarket
	if supplyMarket == nil || supplyMarket.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	borrowMarket := contextSnapshot.BorrowMarket
	if borrowMarket == nil || borrowMarket.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	supply := contextSnapshot.Supply
	if supply == nil || supply.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeLiquidate, core.ErrSupplyNotFound)
	}

	borrow := contextSnapshot.Borrow
	if borrow == nil || borrow.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeLiquidate, core.ErrBorrowNotFound)
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

	borrowPrice := borrowMarket.Price
	if borrowPrice.LessThanOrEqual(decimal.Zero) {
		log.Errorln(e)
		return e
	}

	supplyPrice := supplyMarket.Price
	if supplyPrice.LessThanOrEqual(decimal.Zero) {
		log.Errorln(e)
		return e
	}

	// refund to liquidator if seize not allowed
	if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, output.CreatedAt) {
		return w.abortTransaction(ctx, tx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrSeizeNotAllowed)
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
	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Sub(seizedCTokens).Truncate(16)
		if e = w.supplyStore.Update(ctx, supply, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	//update supply market ctokens
	if e = w.marketStore.Update(ctx, supplyMarket, output.ID); e != nil {
		log.Errorln(e)
		return e
	}
	// supply market transaction
	supplyMarketTransaction := core.BuildMarketUpdateTransaction(ctx, supplyMarket, foxuuid.Modify(output.TraceID, "update_supply_market"))
	if e = w.transactionStore.Create(ctx, supplyMarketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
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
	if output.ID > borrow.Version {
		borrow.Principal = newBorrowBalance.Truncate(16)
		borrow.InterestIndex = newIndex.Truncate(16)
		if e = w.borrowStore.Update(ctx, borrow, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// update borrow market
	if output.ID > borrowMarket.Version {
		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Sub(reallyRepayAmount).Truncate(16)
		borrowMarket.TotalCash = borrowMarket.TotalCash.Add(reallyRepayAmount).Truncate(16)
		if e = w.marketStore.Update(ctx, borrowMarket, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// borrow market transaction
	borrowMarketTransaction := core.BuildMarketUpdateTransaction(ctx, borrowMarket, foxuuid.Modify(output.TraceID, "update_borrow_market"))
	if e = w.transactionStore.Create(ctx, borrowMarketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
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

	extra.Put(core.TransactionKeySupply, core.ExtraSupply{
		UserID:        seizedUserID,
		CTokenAssetID: supply.CTokenAssetID,
		Collaterals:   supply.Collaterals,
	})
	extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
		UserID:        seizedUserID,
		AssetID:       borrow.AssetID,
		Principal:     borrow.Principal,
		InterestIndex: borrow.InterestIndex,
	})

	// liquidation transaction
	tx.SetExtraData(extra)
	tx.Status = core.TransactionStatusComplete
	if e = w.transactionStore.Update(ctx, tx); e != nil {
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
