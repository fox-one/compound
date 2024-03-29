package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"

	"github.com/fox-one/pkg/logger"
)

// handle pledge event
func (w *Payee) handlePledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "pledge")

	market, err := w.mustGetMarketWithCToken(ctx, output.AssetID)
	if err != nil {
		return err
	}

	if market.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	if err := compound.Require(!market.IsMarketClosed(), "payee/market-closed", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("market closed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypePledge, core.ErrMarketClosed)
	}

	//accrue interest
	AccrueInterest(ctx, market, output.CreatedAt)

	supply, err := w.getOrCreateSupply(ctx, userID, output.AssetID)
	if err != nil {
		return err
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.FindByTraceID")
		return err
	}

	if tx.ID == 0 {
		if err := compound.Require(
			output.Amount.LessThanOrEqual(market.CTokens) && market.CollateralFactor.IsPositive(),
			"payee/pledge-denied",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Infoln("pledge denied")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypePledge, core.ErrPledgeNotAllowed)
		}

		totalPledge, err := w.supplyStore.SumOfSupplies(ctx, market.CTokenAssetID)
		if err != nil {
			log.WithError(err).Errorln("supplies.SumOfSupplies")
			return err
		}

		if err := compound.Require(
			!market.MaxPledge.IsPositive() || totalPledge.Add(output.Amount).LessThanOrEqual(market.MaxPledge),
			"payee/max-pledge-exceeded",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Errorln("refund: pledge exceed")
			return err
		}

		extra := core.NewTransactionExtra()
		extra.Put("ctoken_asset_id", output.AssetID)
		extra.Put("amount", output.Amount)
		{
			// useless...
			newCollaterals := supply.Collaterals.Add(output.Amount)
			extra.Put("new_collaterals", newCollaterals)
			extra.Put(core.TransactionKeySupply, core.ExtraSupply{
				UserID:        userID,
				CTokenAssetID: output.AssetID,
				Collaterals:   newCollaterals,
			})
		}
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypePledge, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			log.WithError(err).Errorln("transactions.Create")
			return err
		}
	}

	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Add(output.Amount)
		if err := w.supplyStore.Update(ctx, supply, output.ID); err != nil {
			log.WithError(err).Errorln("supplies.Update")
			return err
		}
	}

	if output.ID > market.Version {
		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("update market error")
			return err
		}
	}

	log.Infoln("pledge completed")
	return nil
}
