package payee

import (
	"compound/core"
	"compound/core/proposal"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handleWithdrawEvent(ctx context.Context, p *core.Proposal, req proposal.WithdrawReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "withdraw",
		"asset":    req.Asset,
		"amount":   req.Amount,
		"opponent": req.Opponent,
	})
	ctx = logger.WithContext(ctx, log)

	amount := req.Amount.Truncate(8)

	// check the market
	market, err := w.marketStore.Find(ctx, req.Asset)
	if err != nil {
		log.WithError(err).Errorln("markets.Find")
		return err
	}
	if market.ID == 0 {
		log.WithError(err).Errorln("skip: market not found")
		return errProposalSkip
	}

	// check the amount
	if amount.GreaterThan(market.Reserves) {
		log.WithField("reserves", market.Reserves).Errorln("insufficient reserves")
		return errProposalSkip
	}

	if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{
		{
			TraceID:   uuid.Modify(p.TraceID, "withdraw-reverses"),
			AssetID:   req.Asset,
			Amount:    amount,
			Threshold: 1,
			Opponents: []string{req.Opponent},
		},
	}); err != nil {
		log.WithError(err).Errorln("wallets.CreateTransfers")
		return err
	}

	if output.ID > market.Version {
		// update market total_cash and reserves
		market.TotalCash = market.TotalCash.Sub(amount)
		market.Reserves = market.Reserves.Sub(amount)

		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("update market error")
			return err
		}
	}

	return nil
}
