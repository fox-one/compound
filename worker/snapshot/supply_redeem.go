package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
)

// handle redeem event
func (w *Payee) handleRedeemEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "supply_redeem")
	ctokenAssetID := output.AssetID

	log.Infof("ctokenAssetID:%s, amount:%s", ctokenAssetID, output.Amount)

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

		cs := core.NewContextSnapshot(nil, nil, supplyMarket, nil)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeRedeem, output, cs)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	contextSnapshot, e := tx.UnmarshalContextSnapshot()
	if e != nil {
		return e
	}

	market := contextSnapshot.SupplyMarket
	if market == nil || market.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeRedeem, core.ErrMarketNotFound)
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeRedeem, core.ErrMarketClosed)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	redeemTokens := output.Amount
	if redeemTokens.GreaterThan(market.CTokens) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeRedeem, core.ErrRedeemNotAllowed)
	}

	// check redeem allowed
	if allowed := w.supplyService.RedeemAllowed(ctx, redeemTokens, market); !allowed {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeRedeem, core.ErrRedeemNotAllowed)
	}

	// transfer asset to user
	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}

	amount := redeemTokens.Mul(exchangeRate).Truncate(8)

	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(amount).Truncate(16)
		market.CTokens = market.CTokens.Sub(redeemTokens).Truncate(16)
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, uuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyAssetID, market.AssetID)
	extra.Put(core.TransactionKeyAmount, amount)
	tx.SetExtraData(extra)
	tx.Status = core.TransactionStatusComplete
	if e = w.transactionStore.Update(ctx, tx); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// transfer
	transferAction := core.TransferAction{
		Source:   core.ActionTypeRedeemTransfer,
		FollowID: followID,
	}
	return w.transferOut(ctx, userID, followID, output.TraceID, market.AssetID, amount, &transferAction)
}
