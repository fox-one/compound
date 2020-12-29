package snapshot

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

func (w *Payee) handlePledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "pledge")
	ctokens := output.Amount
	ctokenAssetID := output.AssetID

	log.Infof("ctokenAssetID:%s, amount:%s", ctokenAssetID, ctokens)

	market, isRecordNotFound, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if isRecordNotFound {
		log.Warningln("market not found")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}
	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if ctokens.GreaterThan(market.CTokens) {
		log.Errorln(errors.New("ctoken overflow"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrPledgeNotAllowed, "")
	}

	if market.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		log.Errorln(errors.New("pledge disallowed"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrPledgeNotAllowed, "")
	}

	return w.db.Tx(func(tx *db.DB) error {
		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, output.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		supply, isRecordNotFound, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
		if e != nil {
			if isRecordNotFound {
				//new
				supply = &core.Supply{
					UserID:        userID,
					CTokenAssetID: ctokenAssetID,
					Collaterals:   ctokens,
				}
				if e = w.supplyStore.Save(ctx, tx, supply); e != nil {
					log.Errorln(e)
					return e
				}
				// add transaction
				transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypePledge, output, nil)
				if e = w.transactionStore.Create(ctx, tx, transaction); e != nil {
					log.WithError(e).Errorln("create transaction error")
					return e
				}
				return nil
			}
			log.Errorln(e)
			return e
		}
		//update supply
		supply.Collaterals = supply.Collaterals.Add(ctokens).Truncate(16)
		e = w.supplyStore.Update(ctx, tx, supply)
		if e != nil {
			log.Errorln(e)
			return e
		}

		// add transaction
		transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypePledge, output, nil)
		if e = w.transactionStore.Create(ctx, tx, transaction); e != nil {
			log.WithError(e).Errorln("create transaction error")
			return e
		}

		return nil
	})
}
