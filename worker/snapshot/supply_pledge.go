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

	market, err := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if err != nil {
		return err
	}
	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrMarketNotFound)
	}

	if market.IsMarketClosed() {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrMarketClosed)
	}

	//accrue interest
	if err := w.marketService.AccrueInterest(ctx, market, output.CreatedAt); err != nil {
		log.WithError(err).Errorln("accrue interest error")
		return err
	}

	supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
	if e != nil {
		return e
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		if ctokens.GreaterThan(market.CTokens) {
			log.Errorln(errors.New("ctoken overflow"))
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrPledgeNotAllowed)
		}

		// check collateral
		if !market.CollateralFactor.IsPositive() {
			log.Errorln(errors.New("pledge disallowed"))
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypePledge, core.ErrPledgeNotAllowed)
		}

		newCollaterals := decimal.Zero
		if supply.ID == 0 {
			newCollaterals = ctokens
		} else {
			newCollaterals = supply.Collaterals.Add(ctokens).Truncate(16)
		}

		extra := core.NewTransactionExtra()
		extra.Put("new_collaterals", newCollaterals)
		extra.Put(core.TransactionKeySupply, core.ExtraSupply{
			UserID:        userID,
			CTokenAssetID: ctokenAssetID,
			Collaterals:   newCollaterals,
		})

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypePledge, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		NewCollaterals decimal.Decimal `json:"new_collaterals"`
	}

	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	if supply.ID == 0 {
		//not exists, create
		supply = &core.Supply{
			UserID:        userID,
			CTokenAssetID: ctokenAssetID,
			Collaterals:   extra.NewCollaterals,
			Version:       output.ID,
		}
		if e = w.supplyStore.Create(ctx, supply); e != nil {
			log.Errorln(e)
			return e
		}
	} else {
		//exists, update supply
		if output.ID > supply.Version {
			supply.Collaterals = extra.NewCollaterals
			if e := w.supplyStore.Update(ctx, supply, output.ID); e != nil {
				return e
			}
		}
	}

	if output.ID > market.Version {
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.WithError(e).Errorln("update market error")
			return e
		}
	}

	return nil
}
