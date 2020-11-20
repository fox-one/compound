package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"
	"strings"
)

var handleRequestSupplyEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	if snapshot.AssetID != w.config.App.GasAssetID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	userID := action[core.ActionKeyUser]
	symbol := strings.ToUpper(action[core.ActionKeySymbol])

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		return e
	}

	supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrSupplyNotFound)
	}

	action = core.NewAction()
	action[core.ActionKeyService] = core.ActionServiceSuppyResponse
	action[core.ActionKeyUser] = userID
	action[core.ActionKeySymbol] = symbol
	action[core.ActionKeyCToken] = supply.Collaterals.String()
	memoStr, e := action.Format()
	if e != nil {
		return e
	}
	trace := id.UUIDFromString(fmt.Sprintf("supply-%s", snapshot.TraceID))
	input := core.Transfer{
		AssetID:    w.config.App.GasAssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     core.GasCost,
		TraceID:    trace,
		Memo:       memoStr,
	}
	if e = w.transferStore.Create(ctx, w.db, &input); e != nil {
		return e
	}

	return nil
}
