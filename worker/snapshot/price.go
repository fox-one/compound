package snapshot

import (
	"compound/core"
	"context"
	"strings"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

var handlePriceEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "price")
	//防止链上价格恶意更改
	if snapshot.OpponentID != w.blockWallet.Client.ClientID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	symbol := strings.ToUpper(action[core.ActionKeySymbol])
	price, e := decimal.NewFromString(action[core.ActionKeyPrice])
	if e != nil {
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidPrice)
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	return w.db.Tx(func(tx *db.DB) error {
		// accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			return e
		}

		market.Price = price.Truncate(8)
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}
		return nil
	})
}
