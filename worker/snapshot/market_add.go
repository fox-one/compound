package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"strings"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
)

func (w *Payee) handleAddMarketEvent(ctx context.Context, p *core.Proposal, req proposal.AddMarketReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithField("worker", "add-market")

	log.Infof("asset:%s", req.AssetID)
	_, isRecordNotFound, e := w.marketStore.Find(ctx, req.AssetID)
	if e == nil {
		//market exists
		return nil
	}

	if isRecordNotFound {
		market := &core.Market{
			Symbol:        strings.ToUpper(req.Symbol),
			AssetID:       req.AssetID,
			CTokenAssetID: req.CTokenAssetID,
			Status:        core.MarketStatusClose,
		}

		if e = w.marketStore.Save(ctx, market); e != nil {
			return e
		}

		// market transaction
		marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
		if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
			log.WithError(e).Errorln("create transaction error")
			return e
		}
	}

	return e
}
