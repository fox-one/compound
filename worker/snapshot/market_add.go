package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"strings"
)

func (w *Payee) handleAddMarketEvent(ctx context.Context, p *core.Proposal, req proposal.AddMarketReq) error {
	_, isRecordNotFound, e := w.marketStore.Find(ctx, req.AssetID)
	if e == nil {
		//market exists
		return nil
	}

	if isRecordNotFound {
		market := core.Market{
			Symbol:        strings.ToUpper(req.Symbol),
			AssetID:       req.AssetID,
			CTokenAssetID: req.CTokenAssetID,
		}

		if e = w.marketStore.Save(ctx, w.db, &market); e != nil {
			return e
		}
	}

	return e
}
