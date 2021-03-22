package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleSupplyEvent(ctx context.Context, tx *db.DB, output *core.Output, userID, followID string, body []byte) error {

	log := logger.FromContext(ctx).WithField("worker", "supply")

	supplyAmount := output.Amount.Abs()
	assetID := output.AssetID

	market, isRecordNotFound, e := w.marketStore.Find(ctx, assetID)
	if isRecordNotFound {
		log.Warningln("market not found")
		return w.handleRefundEvent(ctx, tx, output, userID, followID, core.ActionTypeSupply, core.ErrMarketNotFound, "")
	}
	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.handleRefundEvent(ctx, tx, output, userID, followID, core.ActionTypeSupply, core.ErrMarketClosed, "")
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, tx, market, output.CreatedAt); e != nil {
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
		return w.handleRefundEvent(ctx, tx, output, userID, followID, core.ActionTypeSupply, core.ErrInvalidAmount, "")
	}

	//update maket
	market.CTokens = market.CTokens.Add(ctokens).Truncate(16)
	market.TotalCash = market.TotalCash.Add(supplyAmount).Truncate(16)
	if e = w.marketStore.Update(ctx, tx, market); e != nil {
		log.Errorln(e)
		return e
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, tx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	// add transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyCTokenAssetID, market.CTokenAssetID)
	extra.Put(core.TransactionKeyAmount, ctokens)
	transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeSupply, output, &extra)
	if e = w.transactionStore.Create(ctx, tx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// add transfer task
	transferAction := core.TransferAction{
		Source:   core.ActionTypeMint,
		FollowID: followID,
	}
	return w.transferOut(ctx, tx, userID, followID, output.TraceID, market.CTokenAssetID, ctokens, &transferAction)
}
