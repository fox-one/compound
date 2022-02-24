package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
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
	if w.HasClosedMarkets(ctx) {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrMarketClosed)
	}

	seizedUser, e := w.userStore.FindByAddress(ctx, seizedAddress.String())
	if e != nil {
		return e
	} else if seizedUser.ID == 0 {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrInvalidArgument)
	}

	seizedUserID := seizedUser.UserID
	seizedCTokenAssetID := seizedCTokenAsset.String()

	userPayAmount := output.Amount
	userPayAssetID := output.AssetID

	log.Infof("seizedUser:%s, seizedAsset:%s, payAsset:%s, payAmount:%s", seizedUserID, seizedCTokenAssetID, userPayAssetID, userPayAmount)

	supplyMarket, e := w.marketStore.FindByCToken(ctx, seizedCTokenAssetID)
	if e != nil {
		return e
	}
	if supplyMarket.ID == 0 {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	borrowMarket, e := w.marketStore.Find(ctx, userPayAssetID)
	if e != nil {
		return e
	}
	if borrowMarket.ID == 0 {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	supply, e := w.supplyStore.Find(ctx, seizedUserID, supplyMarket.CTokenAssetID)
	if e != nil {
		return e
	}
	if supply.ID == 0 {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrSupplyNotFound)
	}

	borrow, e := w.borrowStore.Find(ctx, seizedUserID, borrowMarket.AssetID)
	if e != nil {
		return e
	}
	if borrow.ID == 0 {
		return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrBorrowNotFound)
	}

	//supply market accrue interest
	if e = AccrueInterest(ctx, supplyMarket, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	//borrow market accrue interest
	if e = AccrueInterest(ctx, borrowMarket, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		supplyExchangeRate := supplyMarket.CurExchangeRate()

		borrowPrice := borrowMarket.Price
		if !borrowPrice.IsPositive() {
			log.Errorln(e)
			return e
		}

		supplyPrice := supplyMarket.Price
		if !supplyPrice.IsPositive() {
			log.Errorln(e)
			return e
		}

		liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, seizedUserID, borrowMarket, supplyMarket)
		if e != nil {
			log.Errorln(e)
			return e
		}

		// refund to liquidator if seize not allowed
		if !w.accountService.SeizeTokenAllowed(ctx, supply, borrow, liquidity) {
			return w.handleRefundEvent(ctx, output, liquidator, followID, core.ActionTypeLiquidate, core.ErrSeizeNotAllowed)
		}

		borrowBalance, e := compound.BorrowBalance(ctx, borrow, borrowMarket)
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

		repayAmount := repayValue.Div(borrowPrice).Truncate(compound.MaxPricision)
		newBorrowBalance := borrowBalance.Sub(repayAmount).Truncate(compound.MaxPricision)
		newIndex := borrowMarket.BorrowIndex
		if !newBorrowBalance.IsPositive() {
			newBorrowBalance = decimal.Zero
			newIndex = decimal.Zero
			repayAmount = borrowBalance
		}
		refundAmount := userPayAmount.Sub(repayAmount).Truncate(8)

		extra := core.NewTransactionExtra()
		extra.Put(core.TransactionKeyCTokenAssetID, seizedCTokenAssetID)
		extra.Put(core.TransactionKeyAmount, seizedCTokens)
		extra.Put(core.TransactionKeyPrice, seizedPrice)
		if refundAmount.GreaterThan(decimal.Zero) {
			extra.Put(core.TransactionKeyRefund, refundAmount)
		} else {
			extra.Put(core.TransactionKeyRefund, decimal.Zero)
		}

		newCollaterals := supply.Collaterals.Sub(seizedCTokens).Truncate(16)
		extra.Put("new_collaterals", newCollaterals)
		extra.Put("new_borrow_balance", newBorrowBalance)
		extra.Put("new_borrow_index", newIndex)
		extra.Put("repay_amount", repayAmount)

		extra.Put(core.TransactionKeySupply, core.ExtraSupply{
			UserID:        seizedUserID,
			CTokenAssetID: supply.CTokenAssetID,
			Collaterals:   newCollaterals,
		})
		extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
			UserID:        seizedUserID,
			AssetID:       borrow.AssetID,
			Principal:     newBorrowBalance,
			InterestIndex: newIndex,
		})
		tx = core.BuildTransactionFromOutput(ctx, liquidator, followID, core.ActionTypeLiquidate, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		SeizedCToken     decimal.Decimal `json:"amount"`
		RefundAmount     decimal.Decimal `json:"refund"`
		NewCollaterals   decimal.Decimal `json:"new_collaterals"`
		NewBorrowBalance decimal.Decimal `json:"new_borrow_balance"`
		NewBorrowIndex   decimal.Decimal `json:"new_borrow_index"`
		RepayAmount      decimal.Decimal `json:"repay_amount"`
	}

	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	//update supply
	if output.ID > supply.Version {
		supply.Collaterals = extra.NewCollaterals
		if e = w.supplyStore.Update(ctx, supply, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// update borrow account
	if output.ID > borrow.Version {
		borrow.Principal = extra.NewBorrowBalance
		borrow.InterestIndex = extra.NewBorrowIndex
		if e = w.borrowStore.Update(ctx, borrow, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// transfer seized ctoken to liquidator
	transferAction := core.TransferAction{
		Source:   core.ActionTypeLiquidateTransfer,
		FollowID: followID,
	}
	if e = w.transferOut(ctx, liquidator, followID, output.TraceID, supplyMarket.CTokenAssetID, extra.SeizedCToken, &transferAction); e != nil {
		return e
	}

	//refund redundant assets to liquidator
	if extra.RefundAmount.GreaterThan(decimal.Zero) {
		refundTransferAction := core.TransferAction{
			Source:   core.ActionTypeLiquidateRefundTransfer,
			FollowID: followID,
		}
		if e = w.transferOut(ctx, liquidator, followID, output.TraceID, output.AssetID, extra.RefundAmount, &refundTransferAction); e != nil {
			return e
		}
	}

	//update supply market ctokens
	if output.ID > supplyMarket.Version {
		if e = w.marketStore.Update(ctx, supplyMarket, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// update borrow market
	if output.ID > borrowMarket.Version {
		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Sub(extra.RepayAmount).Truncate(compound.MaxPricision)
		borrowMarket.TotalCash = borrowMarket.TotalCash.Add(extra.RepayAmount).Truncate(compound.MaxPricision)
		switch w.sysversion {
		case 0:
		default:
			if borrowMarket.TotalBorrows.IsNegative() {
				borrowMarket.TotalBorrows = decimal.Zero
			}
		}
		if e = w.marketStore.Update(ctx, borrowMarket, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	return nil
}

func (w *Payee) HasClosedMarkets(ctx context.Context) bool {
	markets, e := w.marketStore.All(ctx)
	if e != nil {
		return false
	}

	has := false
	for _, m := range markets {
		if m.Status == core.MarketStatusClose {
			has = true
			break
		}
	}

	return has
}
