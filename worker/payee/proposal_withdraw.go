package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/compound"
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
	market, err := w.mustGetMarket(ctx, req.Asset)
	if err != nil {
		log.WithError(err).Errorln("requireMarket")
		return err
	}

	if err := compound.Require(amount.LessThanOrEqual(market.Reserves), "payee/skip/insufficient-reserves"); err != nil {
		log.WithError(err).Errorln("insufficient reserves")
		return err
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
		AccrueInterest(ctx, market, output.CreatedAt)

		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("update market error")
			return err
		}
	}

	log.Infoln("withdraw completed")
	return nil
}
