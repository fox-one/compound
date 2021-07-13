package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// handle supply event
func (w *Payee) handleSupplyEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "supply")

	supplyAmount := output.Amount
	assetID := output.AssetID

	market, e := w.marketStore.Find(ctx, assetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeSupply, core.ErrMarketNotFound)
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeSupply, core.ErrMarketClosed)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
		if e != nil {
			log.Errorln(e)
			return e
		}

		ctokens := supplyAmount.Div(exchangeRate).Truncate(8)
		if ctokens.IsZero() {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeSupply, core.ErrInvalidAmount)
		}

		extra := core.NewTransactionExtra()
		extra.Put(core.TransactionKeyCTokenAssetID, market.CTokenAssetID)
		extra.Put(core.TransactionKeyAmount, ctokens)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeSupply, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		CTokens decimal.Decimal `json:"amount"`
	}

	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	// add transfer task
	transferAction := core.TransferAction{
		Source:   core.ActionTypeMint,
		FollowID: followID,
	}
	if err := w.transferOut(ctx, userID, followID, output.TraceID, market.CTokenAssetID, extra.CTokens, &transferAction); err != nil {
		return err
	}

	//update maket
	if output.ID > market.Version {
		market.CTokens = market.CTokens.Add(extra.CTokens).Truncate(16)
		market.TotalCash = market.TotalCash.Add(supplyAmount).Truncate(16)
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	return nil
}
