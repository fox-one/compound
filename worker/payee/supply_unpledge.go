package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"compound/pkg/mtg"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// handle unpledge event
func (w *Payee) handleUnpledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "unpledge")

	var ctokenAsset uuid.UUID
	var unpledgedAmount decimal.Decimal

	if _, err := mtg.Scan(body, &ctokenAsset, &unpledgedAmount); err != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeUnpledge, core.ErrInvalidArgument)
	}

	log.Infof("ctokenAssetID:%s, amount:%s", ctokenAsset.String(), unpledgedAmount)
	unpledgedAmount = unpledgedAmount.Truncate(8)
	ctokenAssetID := ctokenAsset.String()

	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeUnpledge, core.ErrMarketNotFound)
	}

	if market.IsMarketClosed() {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeUnpledge, core.ErrMarketClosed)
	}

	//accrue interest
	if e = compound.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
	if e != nil {
		return e
	}

	if supply.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeUnpledge, core.ErrSupplyNotFound)
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		if unpledgedAmount.GreaterThan(supply.Collaterals) {
			log.Errorln(errors.New("insufficient collaterals"))
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeUnpledge, core.ErrInsufficientCollaterals)
		}

		// check liqudity
		liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID, market)
		if e != nil {
			log.Errorln(e)
			return e
		}

		price := market.Price
		exchangeRate := market.ExchangeRate
		unpledgedTokenLiquidity := unpledgedAmount.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
		if unpledgedTokenLiquidity.GreaterThan(liquidity) {
			log.Errorf("insufficient liquidity, liquidity:%v, changed_liquidity:%v", liquidity, unpledgedTokenLiquidity)
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeUnpledge, core.ErrInsufficientLiquidity)
		}

		newCollaterals := supply.Collaterals.Sub(unpledgedAmount).Truncate(16)
		extra := core.NewTransactionExtra()
		extra.Put("new_collaterals", newCollaterals)
		extra.Put(core.TransactionKeyCTokenAssetID, ctokenAssetID)
		extra.Put(core.TransactionKeyAmount, unpledgedAmount)
		extra.Put(core.TransactionKeySupply, core.ExtraSupply{
			UserID:        userID,
			CTokenAssetID: ctokenAssetID,
			Collaterals:   newCollaterals,
		})

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeUnpledge, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		NewCollaterals decimal.Decimal `json:"new_collaterals"`
	}
	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	if output.ID > supply.Version {
		supply.Collaterals = extra.NewCollaterals
		if e = w.supplyStore.Update(ctx, supply, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// add transfer
	transferAction := core.TransferAction{
		Source:   core.ActionTypeUnpledgeTransfer,
		FollowID: followID,
	}
	if err := w.transferOut(ctx, userID, followID, output.TraceID, market.CTokenAssetID, unpledgedAmount, &transferAction); err != nil {
		return err
	}

	if output.ID > market.Version {
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.WithError(e).Errorln("update market error")
			return e
		}
	}

	return nil
}
