package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
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
	market, isRecordNotFound, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if isRecordNotFound {
		log.Warningln("market not found")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketNotFound)
	}
	if e != nil {
		log.WithError(e).Errorln("find market error")
		return e
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketClosed)
	}

	if redeemTokens.GreaterThan(market.CTokens) {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrRedeemNotAllowed)
	}

	// supply
	supply, isRecordNotFound, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if isRecordNotFound {
		log.Warningln("supply not found")
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrSupplyNotFound)
	}
	if e != nil {
		log.WithError(e).Errorln("find supply error")
		return e
	}

	if redeemTokens.GreaterThan(supply.Collaterals) {
		log.Errorln(errors.New("insufficient collaterals"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientCollaterals)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	// check redeem allowed
	if allowed := w.supplyService.RedeemAllowed(ctx, redeemTokens, market); !allowed {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrRedeemNotAllowed)
	}

	// check liqudity
	liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		log.Errorln(e)
		return e
	}

	price := market.Price
	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}
	unpledgedTokenLiquidity := redeemTokens.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
	if unpledgedTokenLiquidity.GreaterThan(liquidity) {
		log.Errorln(errors.New("insufficient liquidity"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientLiquidity)
	}

	// update supply
	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Sub(redeemTokens).Truncate(16)
		if e = w.supplyStore.Update(ctx, supply, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// update market
	underlyingAmount := redeemTokens.Mul(exchangeRate).Truncate(8)
	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(underlyingAmount).Truncate(16)
		market.CTokens = market.CTokens.Sub(redeemTokens).Truncate(16)
		if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
			log.Errorln(e)
			return e
		}
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// transaction
	extra := core.NewTransactionExtra()
	extra.Put(core.TransactionKeyAssetID, market.AssetID)
	extra.Put(core.TransactionKeyAmount, underlyingAmount)
	extra.Put(core.TransactionKeyCTokenAssetID, ctokenAssetID)
	extra.Put(core.TransactionKeyCTokens, redeemTokens)
	extra.Put(core.TransactionKeySupply, core.ExtraSupply{
		UserID:        supply.UserID,
		CTokenAssetID: supply.CTokenAssetID,
		Collaterals:   supply.Collaterals,
	})
	transaction := core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickRedeem, output, &extra)
	if e = w.transactionStore.Create(ctx, transaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	// transfer underlying asset
	transferAction := core.TransferAction{
		Source:   core.ActionTypeQuickRedeemTransfer,
		FollowID: followID,
	}
	return w.transferOut(ctx, userID, followID, output.TraceID, market.AssetID, underlyingAmount, &transferAction)
}
