package payee

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// handle redeem event
func (w *Payee) handleRedeemEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "supply_redeem")
	ctokenAssetID := output.AssetID
	redeemTokens := output.Amount

	log.Infof("ctokenAssetID:%s, amount:%s", ctokenAssetID, output.Amount)

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRedeem, core.ErrMarketNotFound)
	}

	if market.IsMarketClosed() {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRedeem, core.ErrMarketClosed)
	}

	//accrue interest
	if e = AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	if tx.ID == 0 {
		if redeemTokens.GreaterThan(market.CTokens) {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRedeem, core.ErrRedeemNotAllowed)
		}

		// check redeem allowed
		if allowed := market.RedeemAllowed(redeemTokens); !allowed {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRedeem, core.ErrRedeemNotAllowed)
		}

		// transfer asset to user
		exchangeRate := market.CurExchangeRate()
		amount := redeemTokens.Mul(exchangeRate).Truncate(8)

		extra := core.NewTransactionExtra()
		extra.Put(core.TransactionKeyAssetID, market.AssetID)
		extra.Put(core.TransactionKeyAmount, amount)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeRedeem, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		Amount decimal.Decimal `json:"amount"`
	}

	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	// transfer
	transferAction := core.TransferAction{
		Source:   core.ActionTypeRedeemTransfer,
		FollowID: followID,
	}
	if err := w.transferOut(ctx, userID, followID, output.TraceID, market.AssetID, extra.Amount, &transferAction); err != nil {
		return err
	}

	// update market
	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(extra.Amount).Truncate(16)
		market.CTokens = market.CTokens.Sub(redeemTokens).Truncate(16)
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	return nil
}
