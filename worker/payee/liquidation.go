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
	log := logger.FromContext(ctx).WithField("event", "liquidation")
	ctx = logger.WithContext(ctx, log)

	var (
		seizedUserID        string
		seizedCTokenAssetID string
	)
	{
		var seizedAddress, seizedCTokenAsset uuid.UUID
		_, err := mtg.Scan(body, &seizedAddress, &seizedCTokenAsset)
		if err := compound.Require(err == nil, "payee/mtgscan"); err != nil {
			log.Infoln("skip: scan memo failed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrInvalidArgument)
		}

		seizedUser, err := w.userStore.FindByAddress(ctx, seizedAddress.String())
		if err != nil {
			log.WithError(err).Errorln("users.FindByAddress")
			return err
		} else if err := compound.Require(seizedUser.ID > 0, "payee/invalid-seized-address"); err != nil {
			log.Infoln("skip: invalid seized address")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrInvalidArgument)
		}
		seizedUserID = seizedUser.UserID
		seizedCTokenAssetID = seizedCTokenAsset.String()
	}

	supplyMarket, err := w.mustGetMarketWithCToken(ctx, seizedCTokenAssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid supply market")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	borrowMarket, err := w.mustGetMarket(ctx, output.AssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid borrow market")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	if borrowMarket.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	//supply market accrue interest
	AccrueInterest(ctx, supplyMarket, output.CreatedAt)
	//borrow market accrue interest
	AccrueInterest(ctx, borrowMarket, output.CreatedAt)

	supply, err := w.mustGetSupply(ctx, seizedUserID, supplyMarket.CTokenAssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid supply")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrSupplyNotFound)
	}

	borrow, err := w.mustGetBorrow(ctx, seizedUserID, borrowMarket.AssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid borrow")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrBorrowNotFound)
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.Find")
		return err
	}

	if tx.ID == 0 {
		if err := w.HasClosedMarkets(ctx, seizedUserID); err != nil {
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketClosed)
		}

		liquidity, err := w.accountService.CalculateAccountLiquidity(ctx, seizedUserID, borrowMarket, supplyMarket)
		if err != nil {
			log.WithError(err).Errorln("accountz.CalculateAccountLiquidity")
			return err
		}

		// refund to liquidator if seize not allowed
		if err := compound.Require(
			w.accountService.SeizeTokenAllowed(ctx, supply, borrow, liquidity),
			"payee/seize-denied",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Infoln("seize denied")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrSeizeNotAllowed)
		}

		supplyExchangeRate := supplyMarket.CurExchangeRate()
		seizedPrice := supplyMarket.Price.Sub(supplyMarket.Price.Mul(supplyMarket.LiquidationIncentive)).Truncate(compound.MaxPricision)
		repayValue := output.Amount.Mul(borrowMarket.Price).Truncate(compound.MaxPricision)
		borrowBalance := compound.BorrowBalance(ctx, borrow, borrowMarket)
		seizedAmount := repayValue.Div(seizedPrice).Truncate(compound.MaxPricision)
		if maxSeizeValue := supply.Collaterals.
			Mul(supplyExchangeRate).
			Mul(supplyMarket.CloseFactor).
			Mul(seizedPrice).
			Truncate(compound.MaxPricision); repayValue.GreaterThan(maxSeizeValue) {

			repayValue = maxSeizeValue
			seizedAmount = repayValue.Div(seizedPrice)
		}

		if borrowBalanceValue := borrowBalance.Mul(borrowMarket.Price).Truncate(compound.MaxPricision); repayValue.GreaterThan(borrowBalanceValue) {
			repayValue = borrowBalanceValue
			seizedAmount = repayValue.Div(seizedPrice)
		}

		seizedCTokens := seizedAmount.Div(supplyExchangeRate).Truncate(8)
		repayAmount := repayValue.Div(borrowMarket.Price).Truncate(compound.MaxPricision)
		if repayAmount.GreaterThan(borrowBalance) {
			repayAmount = borrowBalance
		}

		extra := core.NewTransactionExtra()
		extra.Put("ctoken_asset_id", seizedCTokenAssetID)
		extra.Put("amount", seizedCTokens)
		extra.Put("repay_amount", repayAmount)
		{
			// useless...
			newCollaterals := supply.Collaterals.Sub(seizedCTokens)
			newBorrowBalance := compound.BorrowBalance(ctx, borrow, borrowMarket).Sub(repayAmount)
			extra.Put("price", seizedPrice)
			extra.Put("new_collaterals", newCollaterals)
			extra.Put("new_borrow_balance", newBorrowBalance)
			extra.Put("new_borrow_index", borrowMarket.BorrowIndex)
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
				InterestIndex: borrowMarket.BorrowIndex,
			})
		}
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeLiquidate, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		SeizedCToken decimal.Decimal `json:"amount"`
		RepayAmount  decimal.Decimal `json:"repay_amount"`
	}

	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	//update supply
	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Sub(extra.SeizedCToken).Truncate(compound.MaxPricision)
		if err := w.supplyStore.Update(ctx, supply, output.ID); err != nil {
			log.WithError(err).Errorln("supplies.Update")
			return err
		}
	}

	// update borrow account
	if output.ID > borrow.Version {
		borrow.Principal = compound.BorrowBalance(ctx, borrow, borrowMarket).Sub(extra.RepayAmount).Truncate(compound.MaxPricision)
		borrow.InterestIndex = borrowMarket.BorrowIndex
		if err := w.borrowStore.Update(ctx, borrow, output.ID); err != nil {
			log.WithError(err).Errorln("borrows.Update")
			return err
		}
	}

	// transfer seized ctoken to liquidator
	if err := w.transferOut(
		ctx,
		userID,
		followID,
		output.TraceID,
		supplyMarket.CTokenAssetID,
		extra.SeizedCToken,
		&core.TransferAction{
			Source:   core.ActionTypeLiquidateTransfer,
			FollowID: followID,
		},
	); err != nil {
		log.WithError(err).Errorln("transferOut")
		return err
	}

	//refund redundant assets to liquidator
	if refundAmount := output.Amount.Sub(extra.RepayAmount); refundAmount.IsPositive() {
		if err := w.transferOut(
			ctx,
			userID,
			followID,
			output.TraceID,
			output.AssetID,
			refundAmount,
			&core.TransferAction{
				Source:   core.ActionTypeLiquidateRefundTransfer,
				FollowID: followID,
			},
		); err != nil {
			log.WithError(err).Errorln("transferOut refund")
			return err
		}
	}

	//update supply market ctokens
	if output.ID > supplyMarket.Version {
		//supply market accrue interest
		AccrueInterest(ctx, supplyMarket, output.CreatedAt)
		if err := w.marketStore.Update(ctx, supplyMarket, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	// update borrow market
	if output.ID > borrowMarket.Version {
		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Sub(extra.RepayAmount).Truncate(compound.MaxPricision)
		borrowMarket.TotalCash = borrowMarket.TotalCash.Add(extra.RepayAmount).Truncate(compound.MaxPricision)
		if w.sysversion > 0 {
			if borrowMarket.TotalBorrows.IsNegative() {
				borrowMarket.TotalBorrows = decimal.Zero
			}
		}
		//borrow market accrue interest
		AccrueInterest(ctx, borrowMarket, output.CreatedAt)
		if err := w.marketStore.Update(ctx, borrowMarket, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	return nil
}

func (w *Payee) HasClosedMarkets(ctx context.Context, user string) error {
	log := logger.FromContext(ctx)

	var closedMarkets = map[string]bool{}
	{
		items, err := w.marketStore.All(ctx)
		if err != nil {
			log.WithError(err).Errorln("markets.All")
			return err
		}

		for _, item := range items {
			if item.Status == core.MarketStatusClose {
				closedMarkets[item.CTokenAssetID] = true
				closedMarkets[item.AssetID] = true
			}
		}
	}

	borrows, err := w.borrowStore.FindByUser(ctx, user)
	if err != nil {
		log.WithError(err).Errorln("borrows.FindByUser")
		return err
	}

	for _, borrow := range borrows {
		_, ok := closedMarkets[borrow.AssetID]
		if err := compound.Require(!ok, "payee/market-closed", compound.FlagRefund); err != nil {
			log.WithError(err).Infoln("failure: borrow market closed", borrow.AssetID)
			return err
		}
	}

	supplies, err := w.supplyStore.FindByUser(ctx, user)
	if err != nil {
		log.WithError(err).Errorln("supplies.FindByUser")
		return err
	}

	for _, supply := range supplies {
		_, ok := closedMarkets[supply.CTokenAssetID]
		if err := compound.Require(!ok, "payee/market-closed", compound.FlagRefund); err != nil {
			log.WithError(err).Infoln("failure: supply market closed", supply.CTokenAssetID)
			return err
		}
	}

	log.Infoln("liquidation completed")
	return nil
}
