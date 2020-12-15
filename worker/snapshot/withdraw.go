package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"

	"github.com/fox-one/pkg/logger"
)

func (w *Payee) handleWithdrawEvent(ctx context.Context, p *core.Proposal, req proposal.WithdrawReq) error {
	log := logger.FromContext(ctx)

	_, isRecordNotFound, e := w.marketStore.Find(ctx, req.Asset)
	if e != nil {
		if isRecordNotFound {
			log.WithError(e).Errorln("no market found:", req.Asset)
			return nil
		}
		return e
	}

	amount := req.Amount.Truncate(8)

	transfer := &core.Transfer{
		TraceID:   p.TraceID,
		AssetID:   req.Asset,
		Amount:    amount,
		Threshold: 1,
		Opponents: []string{req.Opponent},
	}

	if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{transfer}); err != nil {
		log.WithError(err).Errorln("wallets.CreateTransfers")
		return err
	}

	return nil
}
