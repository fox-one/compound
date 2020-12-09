package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
)

func (w *Payee) handleRedeemEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "supply_redeem")
	ctokenAssetID := output.AssetID

	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, output.UpdatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	redeemTokens := output.Amount.Abs()
	if redeemTokens.GreaterThan(market.CTokens) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrRedeemNotAllowed, "")
	}

	// check redeem allowed
	allowed := w.supplyService.RedeemAllowed(ctx, redeemTokens, market)
	if !allowed {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrRedeemNotAllowed, "")
	}

	// transfer asset to user
	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}

	amount := redeemTokens.Mul(exchangeRate).Truncate(8)

	return w.db.Tx(func(tx *db.DB) error {
		market.TotalCash = market.TotalCash.Sub(amount)
		market.CTokens = market.CTokens.Sub(redeemTokens)
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, output.UpdatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		transferAction := core.TransferAction{
			Source:        core.ActionTypeRedeemTransfer,
			TransactionID: followID,
		}

		return w.transferOut(ctx, userID, followID, output.TraceID, market.AssetID, amount, &transferAction)
	})
	return nil
}
