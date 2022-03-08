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
		seizedAddress     uuid.UUID
		seizedCTokenAsset uuid.UUID
	)

	{
		_, err := mtg.Scan(body, &seizedAddress, &seizedCTokenAsset)
		if err := compound.Require(err == nil, "payee/skip/mtgscan", compound.FlagNoisy); err != nil {
			log.Infoln("skip: scan memo failed")
			return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrInvalidArgument)
		}
	}

	seizedUser, err := w.userStore.FindByAddress(ctx, seizedAddress.String())
	if err != nil {
		log.WithError(err).Errorln("users.FindByAddress")
		return err
	} else if err := compound.Require(seizedUser.ID > 0, "payee/skip/invalid-seized-address", compound.FlagNoisy); err != nil {
		log.Infoln("skip: invalid seized address")
		return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrInvalidArgument)
	}

	supplyMarket, err := w.requireMarket(ctx, seizedCTokenAsset.String())
	if err != nil {
		log.WithError(err).Infoln("invalid supply market")
		return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	borrowMarket, err := w.requireMarket(ctx, output.AssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid borrow market")
		return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketNotFound)
	}

	if borrowMarket.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	//supply market accrue interest
	AccrueInterest(ctx, supplyMarket, output.CreatedAt)
	//borrow market accrue interest
	AccrueInterest(ctx, borrowMarket, output.CreatedAt)

	supply, err := w.requireSupply(ctx, seizedUser.UserID, supplyMarket.CTokenAssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid supply")
		return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrSupplyNotFound)
	}

	borrow, err := w.requireBorrow(ctx, seizedUser.UserID, borrowMarket.AssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid borrow")
		return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrBorrowNotFound)
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.Find")
		return err
	}

	if tx.ID == 0 {
		if marketClosed, err := w.HasClosedMarkets(ctx, seizedUser.UserID); err != nil {
			log.WithError(err).Errorln("HasClosedMarkets")
			return err
		} else if err := compound.Require(marketClosed, "payee/refund/market-closed", compound.FlagRefund); err != nil {
			log.WithError(err).Infoln("market closed")
			return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrMarketClosed)
		}

		liquidity, err := w.accountService.CalculateAccountLiquidity(ctx, seizedUser.UserID, borrowMarket, supplyMarket)
		if err != nil {
			log.WithError(err).Errorln("accountz.CalculateAccountLiquidity")
			return err
		}

		// refund to liquidator if seize not allowed
		if err := compound.Require(
			!w.accountService.SeizeTokenAllowed(ctx, supply, borrow, liquidity),
			"payee/refund/seize-denied",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Infoln("market closed")
			return w.handleRefundError(ctx, err, output, userID, followID, core.ActionTypeLiquidate, core.ErrSeizeNotAllowed)
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
		extra.Put("ctoken_asset_id", seizedCTokenAsset.String())
		extra.Put("amount", seizedCTokens)
		extra.Put("repay_amount", repayAmount)

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

func (w *Payee) HasClosedMarkets(ctx context.Context, user string) (bool, error) {
	var closedMarkets = map[string]bool{}
	{
		items, err := w.marketStore.All(ctx)
		if err != nil {
			return false, err
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
		return false, err
	}

	for _, borrow := range borrows {
		if _, ok := closedMarkets[borrow.AssetID]; ok {
			return true, nil
		}
	}

	supplies, err := w.supplyStore.FindByUser(ctx, user)
	if err != nil {
		return false, err
	}

	for _, supply := range supplies {
		if _, ok := closedMarkets[supply.CTokenAssetID]; ok {
			return true, nil
		}
	}

	return false, nil
}
