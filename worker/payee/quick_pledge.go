package payee

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// handle quick pledge event, supply and then pledge
func (w *Payee) handleQuickPledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "quick_pledge")

	//supply
	supplyAmount := output.Amount
	assetID := output.AssetID

	market, e := w.marketStore.Find(ctx, assetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrMarketNotFound)
	}

	if market.IsMarketClosed() {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrMarketClosed)
	}

	if !market.CollateralFactor.IsPositive() {
		log.Errorln(errors.New("pledge disallowed"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrPledgeNotAllowed)
	}

	supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		return e
	}

	//accrue interest
	if e = AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		exchangeRate := market.CurExchangeRate()

		ctokens := supplyAmount.Div(exchangeRate).Truncate(8)
		if ctokens.IsZero() {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickPledge, core.ErrInvalidAmount)
		}

		newCollaterals := decimal.Zero
		if supply.ID == 0 {
			newCollaterals = ctokens
		} else {
			newCollaterals = supply.Collaterals.Add(ctokens).Truncate(16)
		}

		extra := core.NewTransactionExtra()
		extra.Put(core.TransactionKeyCTokenAssetID, market.CTokenAssetID)
		extra.Put(core.TransactionKeyAmount, ctokens)
		extra.Put(core.TransactionKeySupply, core.ExtraSupply{
			UserID:        userID,
			CTokenAssetID: market.CTokenAssetID,
			Collaterals:   newCollaterals,
		})

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickPledge, output, extra)
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

	// pledge
	if supply.ID == 0 {
		//not exists, create
		supply = &core.Supply{
			UserID:        userID,
			CTokenAssetID: market.CTokenAssetID,
			Collaterals:   extra.CTokens,
			Version:       output.ID,
		}
		if e = w.supplyStore.Create(ctx, supply); e != nil {
			log.Errorln(e)
			return e
		}
	} else {
		//exists, update supply
		if output.ID > supply.Version {
			supply.Collaterals = supply.Collaterals.Add(extra.CTokens).Truncate(16)
			e = w.supplyStore.Update(ctx, supply, output.ID)
			if e != nil {
				log.Errorln(e)
				return e
			}
		}
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
