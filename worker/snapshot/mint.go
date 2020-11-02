package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
)

// send ctoken to user
func (w *Worker) handleMintEvent(ctx context.Context, snapshot *core.Snapshot) error {
	return w.db.Tx(func(tx *db.DB) error {
		market, e := w.marketStore.FindByCToken(ctx, snapshot.AssetID, "")
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
		supplies, e := w.supplyStore.Find(ctx, snapshot.OpponentID, market.Symbol)
		if e != nil {
			return e
		}

		if len(supplies) <= 0 {
			//new
			supply := core.Supply{
				UserID:  snapshot.OpponentID,
				Symbol:  market.Symbol,
				CTokens: ctokens,
			}
			e = w.supplyStore.Save(ctx, &supply)
			if e != nil {
				return e
			}
		} else {
			//update
			s := supplies[0]
			s.CTokens = s.CTokens.Add(ctokens)
			e := w.supplyStore.Update(ctx, tx, s)
			if e != nil {
				return e
			}
		}

		return nil
	})

}
