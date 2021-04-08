package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
)

func (w *Payee) handleWithdrawEvent(ctx context.Context, p *core.Proposal, req proposal.WithdrawReq, output *core.Output) error {
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

	// update market total_cash and reserves
	market.TotalCash = market.TotalCash.Sub(amount)
	market.Reserves = market.Reserves.Sub(amount)

	if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
		log.WithError(err).Errorln("update market error")
		return err
	}

	// market transaction
	transaction := core.BuildMarketUpdateTransaction(ctx, market, output.TraceID)
	if e = w.transactionStore.Create(ctx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	transfer, err := core.NewTransfer(p.TraceID, req.Asset, amount, req.Opponent)
	if err != nil {
		log.WithError(err).Errorln("new transfer error")
		return err
	}

	if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{transfer}); err != nil {
		log.WithError(err).Errorln("wallets.CreateTransfers")
		return err
	}

	return nil
}
