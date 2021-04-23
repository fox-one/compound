package snapshot

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

// handle quick pledge event, supply and then pledge
func (w *Payee) handleQuickPledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "quick_pledge")

	//supply
	supplyAmount := output.Amount
	assetID := output.AssetID

	market, isRecordNotFound, e := w.marketStore.Find(ctx, assetID)
	if isRecordNotFound {
		log.Warningln("market not found")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrMarketNotFound)
	}
	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrMarketClosed)
	}

	if market.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		log.Errorln(errors.New("pledge disallowed"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrPledgeNotAllowed)
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
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrInvalidAmount)
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

	// pledge
	supply, isRecordNotFound, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		if isRecordNotFound {
			//not exists, create
			supply = &core.Supply{
				UserID:        userID,
				CTokenAssetID: market.CTokenAssetID,
				Collaterals:   ctokens,
			}
			if e = w.supplyStore.Save(ctx, supply); e != nil {
				log.Errorln(e)
				return e
			}
		} else {
			log.Errorln(e)
			return e
		}
	} else {
		//exists, update supply
		if output.ID > supply.Version {
			supply.Collaterals = supply.Collaterals.Add(ctokens).Truncate(16)
			e = w.supplyStore.Update(ctx, supply, output.ID)
			if e != nil {
				log.Errorln(e)
				return e
			}
		}
	}

	// pledge transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyCTokenAssetID, market.CTokenAssetID)
	extra.Put(core.TransactionKeyAmount, ctokens)
	extra.Put(core.TransactionKeySupply, core.ExtraSupply{
		UserID:        supply.UserID,
		CTokenAssetID: supply.CTokenAssetID,
		Collaterals:   supply.Collaterals,
	})
	transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickPledge, output, extra)
	if e = w.transactionStore.Create(ctx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	return nil
}
