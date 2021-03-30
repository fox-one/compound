package snapshot

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// handle pledge event
func (w *Payee) handlePledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "pledge")
	ctokens := output.Amount
	ctokenAssetID := output.AssetID

	log.Infof("ctokenAssetID:%s, amount:%s", ctokenAssetID, ctokens)

	market, isRecordNotFound, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if isRecordNotFound {
		log.Warningln("market not found")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrMarketNotFound)
	}
	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrMarketClosed)
	}

	if ctokens.GreaterThan(market.CTokens) {
		log.Errorln(errors.New("ctoken overflow"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrPledgeNotAllowed)
	}

	if market.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		log.Errorln(errors.New("pledge disallowed"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrPledgeNotAllowed)
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

	supply, isRecordNotFound, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
	if e != nil {
		if isRecordNotFound {
			//not exists, create
			supply = &core.Supply{
				UserID:        userID,
				CTokenAssetID: ctokenAssetID,
				Collaterals:   ctokens,
			}
			if e = w.supplyStore.Save(ctx, supply); e != nil {
				log.Errorln(e)
				return e
			}
			// add transaction
			transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypePledge, output, nil)
			if e = w.transactionStore.Create(ctx, transaction); e != nil {
				log.WithError(e).Errorln("create transaction error")
				return e
			}
			return nil
		}
		log.Errorln(e)
		return e
	}
	//exists, update supply
	supply.Collaterals = supply.Collaterals.Add(ctokens).Truncate(16)
	e = w.supplyStore.Update(ctx, supply, output.ID)
	if e != nil {
		log.Errorln(e)
		return e
	}

	// add transaction
	transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypePledge, output, nil)
	if e = w.transactionStore.Create(ctx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	return nil
}
