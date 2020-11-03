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

	//utilization rate
	utilizationRate, err := decimal.NewFromString(action[core.ActionKeyUtilizationRate])
	if err != nil {
		return err
	}

	w.marketService.SaveUtilizationRate(ctx, symbol, utilizationRate, block)

	// borrow rate
	borrowRate, err := decimal.NewFromString(action[core.ActionKeyBorrowRate])
	if err != nil {
		return err
	}

	w.marketService.SaveBorrowRatePerBlock(ctx, symbol, borrowRate, block)

	// supply rate
	supplyRate, err := decimal.NewFromString(action[core.ActionKeySupplyRate])
	if err != nil {
		return err
	}
	w.marketService.SaveSupplyRatePerBlock(ctx, symbol, supplyRate, block)

	return nil
}
