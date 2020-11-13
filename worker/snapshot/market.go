package snapshot

import (
	"compound/core"
	"context"
	"strconv"

	"github.com/shopspring/decimal"
)

var handleMarketEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	block, err := strconv.ParseInt(action[core.ActionKeyBlock], 10, 64)
	if err != nil {
		return err
	}
	symbol := action[core.ActionKeySymbol]

	// borrow rate
	borrowRate, err := decimal.NewFromString(action[core.ActionKeyBorrowRate])
	if err != nil {
		return err
	}

	w.marketService.UpdateBorrowRatePerBlock(ctx, symbol, borrowRate, block)

	// supply rate
	supplyRate, err := decimal.NewFromString(action[core.ActionKeySupplyRate])
	if err != nil {
		return err
	}
	w.marketService.UpdateSupplyRatePerBlock(ctx, symbol, supplyRate, block)

	return nil
}
