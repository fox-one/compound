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

	market, err := w.marketStore.Find(ctx, req.AssetID)
	if err != nil {
		log.WithError(err).Errorln("markets.Find")
		return err
	}

	if market.ID == 0 {
		log.WithError(err).Errorln("skip: market not found")
		return errProposalSkip
	}

	if err := AccrueInterest(ctx, market, output.CreatedAt); err != nil {
		log.WithError(err).Errorln("AccrueInterest")
		return err
	}

	market.Status = core.MarketStatusOpen
	if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
		log.WithError(err).Errorln("markets.Update")
		return err
	}

	return nil
}

func (w *Payee) handleCloseMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketStatusReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "close-market",
		"asset":    req.AssetID,
	})

	market, err := w.marketStore.Find(ctx, req.AssetID)
	if err != nil {
		log.WithError(err).Errorln("markets.Find")
		return err
	}

	if market.ID == 0 {
		log.WithError(err).Errorln("skip: market not found")
		return errProposalSkip
	}

	if err := AccrueInterest(ctx, market, output.CreatedAt); err != nil {
		log.WithError(err).Errorln("AccrueInterest")
		return err
	}

	market.Status = core.MarketStatusClose
	if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
		log.WithError(err).Errorln("markets.Update")
		return err
	}

	return nil
}
