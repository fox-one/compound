package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
)

func (w *Payee) handleWithdrawEvent(ctx context.Context, p *core.Proposal, req proposal.WithdrawReq) error {
	log := logger.FromContext(ctx).WithField("worker", "withdraw")

	amount := req.Amount.Truncate(8)

	// check the market
	market, isRecordNotFound, e := w.marketStore.Find(ctx, req.Asset)
	if isRecordNotFound {
		log.Errorln(errors.New("invalid market"))
		return nil
	}

	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	// check the amount
	if amount.GreaterThan(market.Reserves) {
		log.Errorln("insufficient reserves")
		return nil
	}

	return w.db.Tx(func(tx *db.DB) error {
		// update market total_cash and reserves
		market.TotalCash = market.TotalCash.Sub(amount)
		market.Reserves = market.Reserves.Sub(amount)

		if err := w.marketStore.Update(ctx, tx, market); err != nil {
			log.WithError(err).Errorln("update market error")
			return err
		}

		// create transfer
		transfer := &core.Transfer{
			TraceID:   p.TraceID,
			AssetID:   req.Asset,
			Amount:    amount,
			Threshold: 1,
			Opponents: []string{req.Opponent},
		}

		if err := w.walletStore.CreateTransfers(ctx, w.db, []*core.Transfer{transfer}); err != nil {
			log.WithError(err).Errorln("wallets.CreateTransfers")
			return err
		}

		return nil
	})
}
