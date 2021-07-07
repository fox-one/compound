package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
)

func (w *Payee) handleOpenMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketStatusReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithField("worker", "open-market")

	market, e := w.marketStore.Find(ctx, req.AssetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return nil
	}

	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
	}

	market.Status = core.MarketStatusOpen
	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	return nil
}

func (w *Payee) handleCloseMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketStatusReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithField("worker", "close-market")

	market, e := w.marketStore.Find(ctx, req.AssetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return nil
	}

	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
	}

	market.Status = core.MarketStatusClose
	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	return nil
}
