package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"

	"github.com/fox-one/pkg/logger"
)

// handle quick pledge event, supply and then pledge
func (w *Payee) handleQuickPledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("event", "quick_pledge")

	market, err := w.mustGetMarket(ctx, output.AssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickPledge, core.ErrMarketNotFound)
	}

	if market.Version >= output.ID {
		return nil
	}
	if err := compound.Require(!market.IsMarketClosed(), "payee/market-closed", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("market closed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickPledge, core.ErrMarketClosed)
	}

	if err := compound.Require(market.CollateralFactor.IsPositive(), "payee/pledge-disallowed", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("refund: pledge-disallowed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickPledge, core.ErrPledgeNotAllowed)
	}
	//accrue interest
	AccrueInterest(ctx, market, output.CreatedAt)

	supply, err := w.getOrCreateSupply(ctx, userID, market.CTokenAssetID)
	if err != nil {
		return err
	}

	ctokens := output.Amount.Div(market.CurExchangeRate()).Truncate(8)
	if err := compound.Require(ctokens.IsPositive(), "payee/ctoken-too-small", compound.FlagRefund); err != nil {
		log.WithError(err).Errorln("refund: ctoken too small")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickPledge, core.ErrInvalidAmount)
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.FindByTraceID")
		return err
	}

	if tx.ID == 0 {
		extra := core.NewTransactionExtra()
		extra.Put("ctoken_asset_id", market.CTokenAssetID)
		extra.Put("amount", ctokens)
		{
			// useless...
			extra.Put(core.TransactionKeySupply, core.ExtraSupply{
				UserID:        userID,
				CTokenAssetID: market.CTokenAssetID,
				Collaterals:   supply.Collaterals.Add(ctokens),
			})
		}
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickPledge, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			log.WithError(err).Errorln("transactions.Create")
			return err
		}
	}

	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Add(ctokens).Truncate(compound.MaxPricision)
		if err := w.supplyStore.Update(ctx, supply, output.ID); err != nil {
			log.WithError(err).Errorln("supplies.Update")
			return err
		}
	}

	//update maket
	if output.ID > market.Version {
		market.CTokens = market.CTokens.Add(ctokens).Truncate(compound.MaxPricision)
		market.TotalCash = market.TotalCash.Add(output.Amount).Truncate(compound.MaxPricision)
		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	log.Infoln("quick pledge completed")
	return nil
}
