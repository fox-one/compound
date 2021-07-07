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

	tx, e := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if e != nil {
		return e
	}

	if tx.ID == 0 {
		supplyMarket, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
		if e != nil {
			log.WithError(e).Errorln("find market error")
			return e
		}

		supply, e := w.supplyStore.Find(ctx, userID, supplyMarket.CTokenAssetID)
		if e != nil {
			return e
		}

		cs := core.NewContextSnapshot(supply, nil, supplyMarket, nil)
		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickRedeem, output, cs)
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
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketNotFound)
	}

	supply := contextSnapshot.Supply
	if supply == nil || market.ID == 0 {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrSupplyNotFound)
	}

	if w.marketService.IsMarketClosed(ctx, market) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketClosed)
	}

	if redeemTokens.GreaterThan(market.CTokens) {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrRedeemNotAllowed)
	}

	if redeemTokens.GreaterThan(supply.Collaterals) {
		log.Errorln(errors.New("insufficient collaterals"))
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientCollaterals)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	// check redeem allowed
	if allowed := w.supplyService.RedeemAllowed(ctx, redeemTokens, market); !allowed {
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrRedeemNotAllowed)
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
		return w.abortTransaction(ctx, tx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientLiquidity)
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

	tx.SetExtraData(extra)
	tx.Status = core.TransactionStatusComplete
	if e = w.transactionStore.Update(ctx, tx); e != nil {
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
