package snapshot

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

func (w *Payee) handlePledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "pledge")
	ctokens := output.Amount
	ctokenAssetID := output.AssetID

	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
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
		if e = w.marketService.AccrueInterest(ctx, tx, market, output.UpdatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
		if e != nil {
			if gorm.IsRecordNotFoundError(e) {
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
				return nil
			}
			log.Errorln(e)
			return e
		}
		//update supply
		supply.Collaterals = supply.Collaterals.Add(ctokens)
		e = w.supplyStore.Update(ctx, tx, supply)
		if e != nil {
			log.Errorln(e)
			return e
		}

		return nil
	})
}
