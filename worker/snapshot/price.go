package snapshot

import (
	"compound/core"
	"context"
	"strconv"

	"github.com/shopspring/decimal"
)

var handlePriceEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	block, err := strconv.ParseInt(action[core.ActionKeyBlock], 10, 64)
	if err != nil {
		return err
	}

	symbol := action[core.ActionKeySymbol]
	price, err := decimal.NewFromString(action[core.ActionKeyPrice])
	if err != nil {
		return err
	}

	w.priceService.Save(ctx, symbol, price, block)
	return nil
}
