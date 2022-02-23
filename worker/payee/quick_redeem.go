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

func (w *Payee) handleQuickRedeemEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "quick_redeem")

	var ctokenAsset uuid.UUID
	var redeemTokens decimal.Decimal

	if _, err := mtg.Scan(body, &ctokenAsset, &redeemTokens); err != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInvalidArgument)
	}

	log.Infof("ctokenAssetID:%s, amount:%s", ctokenAsset.String(), redeemTokens)
	redeemTokens = redeemTokens.Truncate(8)
	ctokenAssetID := ctokenAsset.String()

	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		return e
	}

	if market.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketNotFound)
	}

	if market.IsMarketClosed() {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketClosed)
	}

	supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		return e
	}
	if supply.ID == 0 {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrSupplyNotFound)
	}

	//accrue interest
	if e = compound.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		if redeemTokens.GreaterThan(market.CTokens) {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrRedeemNotAllowed)
		}

		if redeemTokens.GreaterThan(supply.Collaterals) {
			log.Errorln(errors.New("insufficient collaterals"))
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientCollaterals)
		}

		// check redeem allowed
		if allowed := compound.RedeemAllowed(ctx, redeemTokens, market); !allowed {
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrRedeemNotAllowed)
		}

		// check liqudity
		liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID, market)
		if e != nil {
			log.Errorln(e)
			return e
		}

		price := market.Price
		exchangeRate, e := compound.CurExchangeRate(ctx, market)
		if e != nil {
			log.Errorln(e)
			return e
		}
		unpledgedTokenLiquidity := redeemTokens.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
		if unpledgedTokenLiquidity.GreaterThan(liquidity) {
			log.Errorf("insufficient liquidity, liquidity:%v, changed_liquidity:%v", liquidity, unpledgedTokenLiquidity)
			return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientLiquidity)
		}

		underlyingAmount := redeemTokens.Mul(exchangeRate).Truncate(8)
		newCollaterals := supply.Collaterals.Sub(redeemTokens).Truncate(16)

		extra := core.NewTransactionExtra()
		extra.Put(core.TransactionKeyAssetID, market.AssetID)
		extra.Put(core.TransactionKeyAmount, underlyingAmount)
		extra.Put(core.TransactionKeyCTokenAssetID, ctokenAssetID)
		extra.Put(core.TransactionKeyCTokens, redeemTokens)
		extra.Put(core.TransactionKeySupply, core.ExtraSupply{
			UserID:        supply.UserID,
			CTokenAssetID: supply.CTokenAssetID,
			Collaterals:   newCollaterals,
		})

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickRedeem, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			return err
		}
	}

	var extra struct {
		Amount decimal.Decimal `json:"amount"`
	}
	if err := tx.UnmarshalExtraData(&extra); err != nil {
		return err
	}

	// update supply
	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Sub(redeemTokens).Truncate(16)
		if e = w.supplyStore.Update(ctx, supply, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// transfer underlying asset
	transferAction := core.TransferAction{
		Source:   core.ActionTypeQuickRedeemTransfer,
		FollowID: followID,
	}
	if err := w.transferOut(ctx, userID, followID, output.TraceID, market.AssetID, extra.Amount, &transferAction); err != nil {
		return err
	}

	// update market
	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(extra.Amount).Truncate(16)
		market.CTokens = market.CTokens.Sub(redeemTokens).Truncate(16)
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	return nil
}
