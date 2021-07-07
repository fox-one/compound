package snapshot

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

// handle pledge event
func (w *Payee) handlePledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "pledge")
	ctokens := output.Amount
	ctokenAssetID := output.AssetID

	log.Infof("ctokenAssetID:%s, amount:%s", ctokenAssetID, ctokens)
	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		supplyMarket, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
		if e != nil {
			log.WithError(e).Errorln("find supply market error")
			return e
		}

		supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
		if e != nil {
			return e
		}

		cs := core.NewContextSnapshot(supply, nil, supplyMarket, nil)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypePledge, output, cs)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	contextSnapshot, e := tx.UnmarshalContextSnapshot()
	if e != nil {
		return e
	}

	market := contextSnapshot.BorrowMarket
	if market == nil || market.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypePledge, core.ErrMarketNotFound)
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypePledge, core.ErrMarketClosed)
	}

	if ctokens.GreaterThan(market.CTokens) {
		log.Errorln(errors.New("ctoken overflow"))
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypePledge, core.ErrPledgeNotAllowed)
	}

	// check collateral
	if market.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		log.Errorln(errors.New("pledge disallowed"))
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypePledge, core.ErrPledgeNotAllowed)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.WithError(e).Errorln("accrue interest error")
		return e
	}

	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.WithError(e).Errorln("update market error")
		return e
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, uuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	supply := contextSnapshot.Supply
	if supply == nil || supply.ID == 0 {
		//not exists, create
		supply = &core.Supply{
			UserID:        userID,
			CTokenAssetID: ctokenAssetID,
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

	// pledge transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeySupply, core.ExtraSupply{
		UserID:        supply.UserID,
		CTokenAssetID: supply.CTokenAssetID,
		Collaterals:   supply.Collaterals,
	})

	tx.SetExtraData(extra)
	tx.Status = core.TransactionStatusComplete
	if e = w.transactionStore.Update(ctx, tx); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	return nil
}
