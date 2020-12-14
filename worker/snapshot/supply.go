package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleSupplyEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "supply")

	supplyAmount := output.Amount.Abs()
	assetID := output.AssetID

	market, e := w.marketStore.Find(ctx, assetID)
	if e != nil {
		//refund to user
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, output.UpdatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}

	ctokens := supplyAmount.Div(exchangeRate).Truncate(8)
	if ctokens.LessThan(decimal.NewFromFloat(0.00000001)) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrInvalidAmount, "")
	}

	return w.db.Tx(func(tx *db.DB) error {
		//update maket
		market.CTokens = market.CTokens.Add(ctokens).Truncate(16)
		market.TotalCash = market.TotalCash.Add(supplyAmount).Truncate(16)
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
			Source:        core.ActionTypeMint,
			TransactionID: followID,
		}

		return w.transferOut(ctx, userID, followID, output.TraceID, assetID, ctokens, &transferAction)
	})
}
