package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// from system send ctoken to user
var handleMintEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return w.db.Tx(func(tx *db.DB) error {
		market, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID)
		if e != nil {
			return e
		}

		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			return e
		}

		// update market ctokens
		supplyAmount, e := decimal.NewFromString(action[core.ActionKeyAmount])
		if e != nil {
			return e
		}

		ctokens := snapshot.Amount.Abs()
		market.CTokens = market.CTokens.Add(ctokens)
		market.TotalCash = market.TotalCash.Add(supplyAmount)
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			return e
		}

		return nil
	})
}
