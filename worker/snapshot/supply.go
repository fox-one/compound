package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

func (w *Worker) handleSupplyEvent(ctx context.Context, snapshot *core.Snapshot) error {
	market, e := w.marketStore.Find(ctx, snapshot.AssetID, "")
	if e != nil {
		return w.handleRefundEvent(ctx, snapshot)
	}

	return w.db.Tx(func(tx *db.DB) error {
		//update market
		exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
		if e != nil {
			return e
		}
		ctokens := snapshot.Amount.Div(exchangeRate)

		market.CTokens = market.CTokens.Add(ctokens)
		e = w.marketStore.Update(ctx, w.db, market)
		if e != nil {
			return e
		}
		//update supply
		

		//mint ctoken
		return nil
	})

	return nil
}
