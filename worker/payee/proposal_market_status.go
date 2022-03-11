package payee

import (
	"compound/core"
	"compound/core/proposal"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handleOpenMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketStatusReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "open-market",
		"asset":    req.AssetID,
	})

	market, err := w.mustGetMarket(ctx, req.AssetID)
	if err != nil {
		log.WithError(err).Errorln("requireMarket")
		return err
	}

	AccrueInterest(ctx, market, output.CreatedAt)
	market.Status = core.MarketStatusOpen
	if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
		log.WithError(err).Errorln("markets.Update")
		return err
	}

	log.Infoln("market opened")
	return nil
}

func (w *Payee) handleCloseMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketStatusReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "close-market",
		"asset":    req.AssetID,
	})

	market, err := w.mustGetMarket(ctx, req.AssetID)
	if err != nil {
		return err
	}

	AccrueInterest(ctx, market, output.CreatedAt)
	market.Status = core.MarketStatusClose
	if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
		log.WithError(err).Errorln("markets.Update")
		return err
	}

	log.Infoln("market closed")
	return nil
}
