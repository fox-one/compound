package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"compound/pkg/mtg"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// handle quick borrow event, supply, then pledge, and then borrow
func (w *Payee) handleQuickBorrowEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("event", "quick_borrow")

	var (
		borrowAssetID string
		borrowAmount  decimal.Decimal
	)
	{
		var asset uuid.UUID
		_, e := mtg.Scan(body, &asset, &borrowAmount)
		if err := compound.Require(e == nil, "payee/mtgscan"); err != nil {
			log.WithError(err).Infoln("skip: scan memo failed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrInvalidArgument)
		}

		borrowAmount = borrowAmount.Truncate(8)
		borrowAssetID = asset.String()
		log = logger.FromContext(ctx).WithFields(logrus.Fields{
			"borrow_asset_id": borrowAssetID,
			"borrow_amount":   borrowAmount,
		})
		ctx = logger.WithContext(ctx, log)
	}

	borrowMarket, err := w.mustGetMarket(ctx, borrowAssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrMarketNotFound)
	}
	if borrowMarket.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	supplyMarket, err := w.mustGetMarket(ctx, output.AssetID)
	if err != nil && errors.As(err, &compound.Error{}) {
		supplyMarket, err = w.mustGetMarketWithCToken(ctx, output.AssetID)
	}
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrMarketNotFound)
	}
	isSupplyCToken := supplyMarket.CTokenAssetID == output.AssetID

	// supply market accrue interest
	AccrueInterest(ctx, supplyMarket, output.CreatedAt)
	//borrow market accrue interest
	AccrueInterest(ctx, borrowMarket, output.CreatedAt)

	if err := compound.Require(
		!supplyMarket.IsMarketClosed() && !borrowMarket.IsMarketClosed(),
		"payee/market-closed",
		compound.FlagRefund,
	); err != nil {
		log.WithError(err).Errorln("refund: market closed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrMarketClosed)
	}

	if err := compound.Require(
		supplyMarket.CollateralFactor.IsPositive(),
		"payee/pledge-disallowed",
		compound.FlagRefund,
	); err != nil {
		log.WithError(err).Errorln("refund: market collateral factor is 0")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrPledgeNotAllowed)
	}

	supply, err := w.getOrCreateSupply(ctx, userID, supplyMarket.CTokenAssetID)
	if err != nil {
		return err
	}

	borrow, err := w.getOrCreateBorrow(ctx, userID, borrowMarket.AssetID)
	if err != nil {
		return err
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.FindByTraceID")
		return err
	}

	if tx.ID == 0 {
		// check liquidity
		liquidity, err := w.accountService.CalculateAccountLiquidity(ctx, userID, supplyMarket, borrowMarket)
		if err != nil {
			log.WithError(err).Errorln("accountz.CalculateAccountLiquidity")
			return err
		}
		// add the additional liquidity provided this time
		if isSupplyCToken {
			liquidity = liquidity.Add(output.Amount.Mul(supplyMarket.ExchangeRate).Mul(supplyMarket.CollateralFactor).Mul(supplyMarket.Price))
		} else {
			liquidity = liquidity.Add(output.Amount.Mul(supplyMarket.CollateralFactor).Mul(supplyMarket.Price))
		}

		borrowableSupplies := borrowMarket.TotalCash.Sub(borrowMarket.Reserves)
		if err := compound.Require(
			borrowAmount.IsPositive() &&
				borrowableSupplies.GreaterThanOrEqual(borrowMarket.BorrowCap) &&
				borrowAmount.LessThanOrEqual(borrowableSupplies.Sub(borrowMarket.BorrowCap)) &&
				borrowAmount.Mul(borrowMarket.Price).LessThanOrEqual(liquidity),
			"payee/borrow-denied",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Errorln("refund: borrow denied")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrBorrowNotAllowed)
		}

		// supply, calculate ctokens
		var ctokens decimal.Decimal
		if isSupplyCToken {
			ctokens = output.Amount
		} else {
			ctokens = output.Amount.Div(supplyMarket.CurExchangeRate()).Truncate(8)
		}

		if err := compound.Require(
			ctokens.IsPositive(),
			"payee/ctokens-too-small",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Errorln("refund: ctokens too small")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrInvalidAmount)
		}

		totalPledge, err := w.supplyStore.SumOfSupplies(ctx, supplyMarket.CTokenAssetID)
		if err != nil {
			log.WithError(err).Errorln("supplies.SumOfSupplies")
			return err
		}

		if err := compound.Require(
			!supplyMarket.MaxPledge.IsPositive() || totalPledge.Add(ctokens).LessThanOrEqual(supplyMarket.MaxPledge),
			"payee/max-pledge-exceeded",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Errorln("refund: pledge exceed")
			return err
		}

		extra := core.NewTransactionExtra()
		extra.Put("asset_id", borrowAssetID)
		extra.Put("amount", borrowAmount)
		extra.Put("ctokens", ctokens)
		extra.Put("ctoken_asset_id", supplyMarket.CTokenAssetID)
		{
			// useless...
			newCollaterals := supply.Collaterals.Add(ctokens)
			newBorrowBalance := compound.BorrowBalance(ctx, borrow, borrowMarket).Add(borrowAmount)
			extra.Put("new_collaterals", newCollaterals)
			extra.Put("new_borrow_balance", newBorrowBalance)
			extra.Put("new_borrow_index", borrowMarket.BorrowIndex)
			extra.Put(core.TransactionKeySupply, core.ExtraSupply{
				UserID:        userID,
				CTokenAssetID: supplyMarket.CTokenAssetID,
				Collaterals:   newCollaterals,
			})
			extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
				UserID:        userID,
				AssetID:       borrowMarket.AssetID,
				Principal:     newBorrowBalance,
				InterestIndex: borrowMarket.BorrowIndex,
			})
		}
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickBorrow, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			log.WithError(err).Errorln("transactions.Create")
			return err
		}
	}

	var extra struct {
		CTokens decimal.Decimal `json:"ctokens"`
	}

	if err := tx.UnmarshalExtraData(&extra); err != nil {
		log.WithError(err).Errorln("transactions.UnmarshalExtraData")
		return err
	}

	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Add(extra.CTokens)
		if err := w.supplyStore.Update(ctx, supply, output.ID); err != nil {
			log.WithError(err).Errorln("supplies.Update")
			return err
		}
	}

	if output.ID > borrow.Version {
		borrow.Principal = compound.BorrowBalance(ctx, borrow, borrowMarket).Add(borrowAmount).Truncate(compound.MaxPricision)
		borrow.InterestIndex = borrowMarket.BorrowIndex
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
		borrowAssetID,
		borrowAmount,
		&core.TransferAction{
			Source:   core.ActionTypeQuickBorrowTransfer,
			FollowID: followID,
		},
	); err != nil {
		return err
	}

	// update supply market
	if output.ID > supplyMarket.Version {
		// Only update the ctokens and total_cash of market when the underlying assets are provided
		if !isSupplyCToken {
			supplyMarket.CTokens = supplyMarket.CTokens.Add(extra.CTokens).Truncate(compound.MaxPricision)
			supplyMarket.TotalCash = supplyMarket.TotalCash.Add(output.Amount).Truncate(compound.MaxPricision)
		}
		AccrueInterest(ctx, supplyMarket, output.CreatedAt)
		if err := w.marketStore.Update(ctx, supplyMarket, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	// update borrow market
	if output.ID > borrowMarket.Version {
		borrowMarket.TotalCash = borrowMarket.TotalCash.Sub(borrowAmount).Truncate(compound.MaxPricision)
		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Add(borrowAmount).Truncate(compound.MaxPricision)
		AccrueInterest(ctx, borrowMarket, output.CreatedAt)
		// update market
		if err := w.marketStore.Update(ctx, borrowMarket, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	log.Infoln("quick borrow completed")
	return nil
}
