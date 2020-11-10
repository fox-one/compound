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
		market, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID, "")
		if e != nil {
			return e
		}

		pAmount := action[core.ActionKeyAmount]
		principal, e := decimal.NewFromString(pAmount)
		if e != nil {
			return e
		}

		ctokens := snapshot.Amount.Abs()

		// update market ctokens
		market.CTokens = market.CTokens.Add(ctokens)
		e = w.marketStore.Update(ctx, tx, market)
		if e != nil {
			return e
		}
		//update supply ctokens
		supply, e := w.supplyStore.Find(ctx, snapshot.OpponentID, market.Symbol)
		if e != nil {
			//new
			supply := core.Supply{
				UserID:    snapshot.OpponentID,
				Symbol:    market.Symbol,
				Principal: principal,
				CTokens:   ctokens,
			}
			e = w.supplyStore.Save(ctx, tx, &supply)
			if e != nil {
				return e
			}
		} else {
			//update
			supply.Principal = supply.Principal.Add(principal)
			supply.CTokens = supply.CTokens.Add(ctokens)
			e := w.supplyStore.Update(ctx, tx, supply)
			if e != nil {
				return e
			}
		}

		return nil
	})

}
