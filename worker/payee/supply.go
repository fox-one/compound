package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"

	"github.com/fox-one/pkg/logger"
)

// handle supply event
func (w *Payee) handleSupplyEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "supply")

	market, err := w.mustGetMarket(ctx, output.AssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeRepay, core.ErrMarketNotFound)
	}

	if market.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	if err := compound.Require(!market.IsMarketClosed(), "payee/market-closed", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("market closed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeSupply, core.ErrMarketClosed)
	}

	//accrue interest
	AccrueInterest(ctx, market, output.CreatedAt)

	ctokens := output.Amount.Div(market.CurExchangeRate()).Truncate(8)
	if err := compound.Require(ctokens.IsPositive(), "payee/amount-too-small", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("skip: amount too small")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeSupply, core.ErrInvalidAmount)
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.Find")
		return err
	}

	if tx.ID == 0 {
		extra := core.NewTransactionExtra()
		extra.Put("ctoken_asset_id", market.CTokenAssetID)
		extra.Put("amount", ctokens)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeSupply, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			log.WithError(err).Errorln("transactions.Create")
			return err
		}
	}

	// add transfer task
	if err := w.transferOut(
		ctx,
		userID,
		followID,
		output.TraceID,
		market.CTokenAssetID,
		ctokens,
		&core.TransferAction{
			Source:   core.ActionTypeMint,
			FollowID: followID,
		},
	); err != nil {
		return err
	}

	//update maket
	if output.ID > market.Version {
		market.CTokens = market.CTokens.Add(ctokens).Truncate(16)
		market.TotalCash = market.TotalCash.Add(output.Amount).Truncate(16)
		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	log.Infoln("supply completed")
	return nil
}
