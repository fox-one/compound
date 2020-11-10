package snapshot

import (
	"compound/core"
	"context"
	"strconv"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// from system, ignore if error
var handleAccountLiquidityEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx)

	block, e := strconv.ParseInt(action[core.ActionKeyBlock], 10, 64)
	if e != nil {
		return nil
	}
	liquidity, e := decimal.NewFromString(action[core.ActionKeyAmount])
	if e != nil {
		return nil
	}
	userID := action[core.ActionKeyUser]

	if e = w.accountStore.SaveLiquidity(ctx, userID, block, liquidity); e != nil {
		log.Errorln("save liquidity error:", e)
		return nil
	}

	return nil
}
