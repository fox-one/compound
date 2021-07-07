package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

// handle supply event
func (w *Payee) handleSupplyEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "supply")

	supplyAmount := output.Amount
	assetID := output.AssetID

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		supplyMarket, e := w.marketStore.Find(ctx, assetID)
		if e != nil {
			log.WithError(e).Errorln("find market error")
			return e
		}

		cs := core.NewContextSnapshot(nil, nil, supplyMarket, nil)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeSupply, output, cs)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	contextSnapshot, e := tx.UnmarshalContextSnapshot()
	if e != nil {
		return e
	}

	market := contextSnapshot.SupplyMarket
	if market == nil || market.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeSupply, core.ErrMarketNotFound)
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeSupply, core.ErrMarketClosed)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}

	ctokens := supplyAmount.Div(exchangeRate).Truncate(8)
	if ctokens.LessThan(decimal.NewFromFloat(0.00000001)) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeSupply, core.ErrInvalidAmount)
	}

	//update maket
	if output.ID > market.Version {
		market.CTokens = market.CTokens.Add(ctokens).Truncate(16)
		market.TotalCash = market.TotalCash.Add(supplyAmount).Truncate(16)
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, uuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyCTokenAssetID, market.CTokenAssetID)
	extra.Put(core.TransactionKeyAmount, ctokens)

	tx.SetExtraData(extra)
	tx.Status = core.TransactionStatusComplete
	if e = w.transactionStore.Update(ctx, tx); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// add transfer task
	transferAction := core.TransferAction{
		Source:   core.ActionTypeMint,
		FollowID: followID,
	}
	return w.transferOut(ctx, userID, followID, output.TraceID, market.CTokenAssetID, ctokens, &transferAction)
}
