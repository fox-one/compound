package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
)

func (w *Payee) handleOpenMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketStatusReq, t time.Time) error {
	return w.db.Tx(func(tx *db.DB) error {
		log := logger.FromContext(ctx).WithField("worker", "open-market")

		market, isRecordNotFound, e := w.marketStore.Find(ctx, req.AssetID)
		if e != nil {
			if isRecordNotFound {
				return nil
			}

			return e
		}

		if e = w.marketService.AccrueInterest(ctx, tx, market, t); e != nil {
			return e
		}

		market.Status = core.MarketStatusOpen
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		return nil
	})
}

func (w *Payee) handleCloseMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketStatusReq, t time.Time) error {
	return w.db.Tx(func(tx *db.DB) error {
		log := logger.FromContext(ctx).WithField("worker", "close-market")

		market, isRecordNotFound, e := w.marketStore.Find(ctx, req.AssetID)
		if e != nil {
			if isRecordNotFound {
				return nil
			}

			return e
		}

		if e = w.marketService.AccrueInterest(ctx, tx, market, t); e != nil {
			return e
		}

		market.Status = core.MarketStatusClose
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		return nil
	})
}
