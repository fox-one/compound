package snapshot

import (
	"compound/core"
	"context"

	"github.com/shopspring/decimal"
)

var handlePriceEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	symbol := action[core.ActionKeySymbol]
	price, e := decimal.NewFromString(action[core.ActionKeyPrice])
	if e != nil {
		return e
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return e
	}
	market.Price = price
	if e = w.marketStore.Update(ctx, w.db, market); e != nil {
		return e
	}

	return nil
}
