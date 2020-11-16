package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// send ctoken to user
var handleMintEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	return w.db.Tx(func(tx *db.DB) error {
		market, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID)
		if e != nil {
			return e
		}

		// update market ctokens
		pAmount := action[core.ActionKeyAmount]
		principal, e := decimal.NewFromString(pAmount)
		if e != nil {
			return e
		}

		ctokens := snapshot.Amount.Abs()
		market.CTokens = market.CTokens.Add(ctokens)
		market.TotalCash = market.TotalCash.Add(principal)

		// keep the flywheel moving
		e = w.marketService.KeppFlywheelMoving(ctx, tx, market, snapshot.CreatedAt)
		if e != nil {
			return e
		}

		//update interest index
		e = w.borrowService.UpdateMarketInterestIndex(ctx, tx, market, market.BlockNumber)
		if e != nil {
			return e
		}

		return nil
	})
}
