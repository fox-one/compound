package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"strings"

	"github.com/jinzhu/gorm"
)

func (w *Payee) handleAddMarketEvent(ctx context.Context, p *core.Proposal, req proposal.AddMarketReq) error {
	_, e := w.marketStore.Find(ctx, req.AssetID)
	if e == nil {
		//market exists
		return nil
	}

	if gorm.IsRecordNotFoundError(e) {
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
