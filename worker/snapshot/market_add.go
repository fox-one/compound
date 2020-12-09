package snapshot

import (
	"compound/core"
	"context"
)

func (w *Payee) handleAddMarketEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	return nil
}

// var handleAddMarketEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
// 	log := logger.FromContext(ctx).WithField("worker", "add-market")
// 	if snapshot.AssetID != w.system.VoteAsset {
// 		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
// 	}

// 	if !w.system.IsAdmin(snapshot.OpponentID) {
// 		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
// 	}

// 	symbol := strings.ToUpper(action[core.ActionKeySymbol])
// 	assetID := action[core.ActionKeyAssetID]
// 	ctokenAssetID := action[core.ActionKeyCTokenAssetID]

// 	_, e := w.marketStore.FindBySymbol(ctx, symbol)
// 	if e == nil {
// 		log.Errorln(e)
// 		// market exists
// 		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
// 	}

// 	market := core.Market{
// 		Symbol:        symbol,
// 		AssetID:       assetID,
// 		CTokenAssetID: ctokenAssetID,
// 	}

// 	if e = w.marketStore.Save(ctx, w.db, &market); e != nil {
// 		log.Errorln(e)
// 		return e
// 	}

// 	return nil
// }
