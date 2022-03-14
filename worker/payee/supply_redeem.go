package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"

	"github.com/fox-one/pkg/logger"
)

// handle redeem event
func (w *Payee) handleRedeemEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("event", "supply_redeem")

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.FindByTraceID")
		return err
	}

	market, err := w.mustGetMarketWithCToken(ctx, output.AssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeRedeem, core.ErrMarketNotFound)
	}

	if market.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	if err := compound.Require(!market.IsMarketClosed(), "payee/market-closed", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("market closed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeRedeem, core.ErrMarketClosed)
	}

	//accrue interest
	AccrueInterest(ctx, market, output.CreatedAt)

	amount := market.CurExchangeRate().Mul(output.Amount).Truncate(8)

	if tx.ID == 0 {
		if err := compound.Require(
			output.Amount.LessThanOrEqual(market.CTokens) && market.RedeemAllowed(output.Amount),
			"payee/pledge-denied",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Infoln("pledge denied")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeRedeem, core.ErrRedeemNotAllowed)
		}

		extra := core.NewTransactionExtra()
		extra.Put("asset_id", market.AssetID)
		extra.Put("amount", amount)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeRedeem, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	if err := w.transferOut(
		ctx,
		userID,
		followID,
		output.TraceID,
		market.AssetID,
		amount,
		&core.TransferAction{
			Source:   core.ActionTypeRedeemTransfer,
			FollowID: followID,
		},
	); err != nil {
		return err
	}

	// update market
	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(amount).Truncate(16)
		market.CTokens = market.CTokens.Sub(output.Amount).Truncate(16)
		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	log.Infoln("redeem completed")
	return nil
}
