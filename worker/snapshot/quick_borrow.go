package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// handle quick borrow event, supply, then pledge, and then borrow
func (w *Payee) handleQuickBorrowEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "quick_borrow")

	// parse params, either underlying asset or ctoken asset
	supplyAmount := output.Amount
	supplyAssetID := output.AssetID

	var borrowAsset uuid.UUID
	var borrowAmount decimal.Decimal
	if _, err := mtg.Scan(body, &borrowAsset, &borrowAmount); err != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrInvalidArgument)
	}

	borrowAmount = borrowAmount.Truncate(8)
	borrowAssetID := borrowAsset.String()

	// check supply market
	isSupplyCToken, e := w.isSupplyCToken(ctx, supplyAssetID)
	if e != nil {
		return e
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		supplyMarket, e := w.marketStore.Find(ctx, supplyAssetID)
		if e != nil {
			log.WithError(e).Errorln("find market error")
			return e
		}
		if supplyMarket.ID == 0 {
			supplyMarket, e = w.marketStore.FindByCToken(ctx, supplyAssetID)
			if e != nil {
				return e
			}
		}

		borrowMarket, e := w.marketStore.Find(ctx, borrowAssetID)
		if e != nil {
			return e
		}

		supply, e := w.supplyStore.Find(ctx, userID, supplyMarket.CTokenAssetID)
		if e != nil {
			return e
		}

		borrow, e := w.borrowStore.Find(ctx, userID, borrowMarket.AssetID)
		if e != nil {
			return e
		}

		cs := core.NewContextSnapshot(supply, borrow, supplyMarket, borrowMarket)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickBorrow, output, cs)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	contextSnapshot, e := tx.UnmarshalContextSnapshot()
	if e != nil {
		return e
	}

	// check supply market
	supplyMarket := contextSnapshot.SupplyMarket
	if supplyMarket == nil || supplyMarket.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrMarketNotFound)
	}

	if w.marketService.IsMarketClosed(ctx, supplyMarket) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrMarketClosed)
	}

	// check collateral
	if supplyMarket.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		log.Errorln(errors.New("pledge disallowed"))
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrPledgeNotAllowed)
	}

	// check borrow market
	borrowMarket := contextSnapshot.BorrowMarket
	if borrowMarket == nil || borrowMarket.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrMarketNotFound)
	}
	if w.marketService.IsMarketClosed(ctx, borrowMarket) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrMarketClosed)
	}

	// supply market accrue interest
	if e = w.marketService.AccrueInterest(ctx, supplyMarket, output.CreatedAt); e != nil {
		return e
	}

	//borrow market accrue interest
	if e = w.marketService.AccrueInterest(ctx, borrowMarket, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	// check borrow ability
	if borrowAmount.LessThanOrEqual(decimal.Zero) {
		log.Errorln("invalid borrow amount")
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrBorrowNotAllowed)
	}

	// check borrow cap
	borrowableSupplies := borrowMarket.TotalCash.Sub(borrowMarket.Reserves)
	if borrowableSupplies.LessThan(borrowMarket.BorrowCap) {
		log.Errorln("insufficient market cash")
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrBorrowNotAllowed)
	}

	if borrowAmount.GreaterThan(borrowableSupplies.Sub(borrowMarket.BorrowCap)) {
		log.Errorln("insufficient market cash")
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrBorrowNotAllowed)
	}

	// check liquidity
	liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		log.Errorln(e)
		return e
	}

	// add the additional liquidity provided this time
	if isSupplyCToken {
		liquidity = liquidity.Add(supplyAmount.Mul(supplyMarket.ExchangeRate).Mul(supplyMarket.CollateralFactor).Mul(supplyMarket.Price))
	} else {
		liquidity = liquidity.Add(supplyAmount.Mul(supplyMarket.CollateralFactor).Mul(supplyMarket.Price))
	}

	borrowValue := borrowAmount.Mul(borrowMarket.Price)
	if borrowValue.GreaterThan(liquidity) {
		log.Errorln("insufficient liquidity")
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrBorrowNotAllowed)
	}

	// supply
	exchangeRate, e := w.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		log.Errorln(e)
		return e
	}

	// supply, calculate ctokens
	ctokens := decimal.Zero
	if isSupplyCToken {
		ctokens = supplyAmount
	} else {
		ctokens = supplyAmount.Div(exchangeRate).Truncate(8)
	}

	if ctokens.LessThan(decimal.NewFromFloat(0.00000001)) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickBorrow, core.ErrInvalidAmount)
	}

	if output.ID > supplyMarket.Version {
		// Only update the ctokens and total_cash of market when the underlying assets are provided
		if !isSupplyCToken {
			supplyMarket.CTokens = supplyMarket.CTokens.Add(ctokens).Truncate(16)
			supplyMarket.TotalCash = supplyMarket.TotalCash.Add(supplyAmount).Truncate(16)
		}

		if e = w.marketStore.Update(ctx, supplyMarket, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// supply market transaction
	supplyMarketTransaction := core.BuildMarketUpdateTransaction(ctx, supplyMarket, foxuuid.Modify(output.TraceID, "update_supply_market"))
	if e = w.transactionStore.Create(ctx, supplyMarketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// update pledge data
	supply := contextSnapshot.Supply
	if supply == nil || supply.ID == 0 {
		//not exists, create
		supply = &core.Supply{
			UserID:        userID,
			CTokenAssetID: supplyMarket.CTokenAssetID,
			Collaterals:   ctokens,
			Version:       output.ID,
		}
		if e = w.supplyStore.Save(ctx, supply); e != nil {
			log.Errorln(e)
			return e
		}
	} else {
		//exists, update supply
		if output.ID > supply.Version {
			supply.Collaterals = supply.Collaterals.Add(ctokens).Truncate(16)
			e = w.supplyStore.Update(ctx, supply, output.ID)
			if e != nil {
				log.Errorln(e)
				return e
			}
		}
	}

	// borrow event
	if output.ID > borrowMarket.Version {
		borrowMarket.TotalCash = borrowMarket.TotalCash.Sub(borrowAmount).Truncate(16)
		borrowMarket.TotalBorrows = borrowMarket.TotalBorrows.Add(borrowAmount).Truncate(16)
		// update market
		if e = w.marketStore.Update(ctx, borrowMarket, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, borrowMarket, foxuuid.Modify(output.TraceID, "update_borrow_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	borrow := contextSnapshot.Borrow
	if borrow == nil || borrow.ID == 0 {
		//new borrow record
		borrow = &core.Borrow{
			UserID:        userID,
			AssetID:       borrowMarket.AssetID,
			Principal:     borrowAmount,
			InterestIndex: borrowMarket.BorrowIndex,
			Version:       output.ID}

		if e = w.borrowStore.Save(ctx, borrow); e != nil {
			log.Errorln(e)
			return e
		}
	} else {
		//update borrow account
		if output.ID > borrow.Version {
			borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, borrowMarket)
			if e != nil {
				log.Errorln(e)
				return e
			}
			newBorrowBalance := borrowBalance.Add(borrowAmount)
			borrow.Principal = newBorrowBalance.Truncate(16)
			borrow.InterestIndex = borrowMarket.BorrowIndex.Truncate(16)
			e = w.borrowStore.Update(ctx, borrow, output.ID)
			if e != nil {
				log.Errorln(e)
				return e
			}
		}
	}

	// transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyAssetID, borrowAssetID)
	extra.Put(core.TransactionKeyAmount, borrowAmount)
	extra.Put(core.TransactionKeySupply, core.ExtraSupply{
		UserID:        supply.UserID,
		CTokenAssetID: supply.CTokenAssetID,
		Collaterals:   supply.Collaterals,
	})
	extra.Put(core.TransactionKeyBorrow, core.ExtraBorrow{
		UserID:        borrow.UserID,
		AssetID:       borrow.AssetID,
		Principal:     borrow.Principal,
		InterestIndex: borrow.InterestIndex,
	})

	tx.SetExtraData(extra)
	tx.Status = core.TransactionStatusComplete
	if e = w.transactionStore.Update(ctx, tx); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	//transfer borrowed asset
	transferAction := core.TransferAction{
		Source:   core.ActionTypeQuickBorrowTransfer,
		FollowID: followID,
	}
	return w.transferOut(ctx, userID, followID, output.TraceID, borrowAssetID, borrowAmount, &transferAction)
}

func (w *Payee) isSupplyCToken(ctx context.Context, supplyAssetID string) (bool, error) {
	isSupplyCToken := false

	supplyMarket, e := w.marketStore.Find(ctx, supplyAssetID)
	if e != nil {
		return isSupplyCToken, e
	}
	if supplyMarket.ID == 0 {
		_, e := w.marketStore.FindByCToken(ctx, supplyAssetID)
		if e != nil {
			return isSupplyCToken, e
		}

		isSupplyCToken = true
	}

	return isSupplyCToken, nil
}
