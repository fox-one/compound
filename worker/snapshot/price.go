package snapshot

import (
	"compound/core"
	"context"
	"strings"

	"github.com/shopspring/decimal"
)

var handlePriceEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	if snapshot.OpponentID != w.blockWallet.Client.ClientID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	symbol := strings.ToUpper(action[core.ActionKeySymbol])
	price, e := decimal.NewFromString(action[core.ActionKeyPrice])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidPrice)
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}
	market.Price = price
	if e = w.marketStore.Update(ctx, w.db, market); e != nil {
		return e
	}

	return nil
}
