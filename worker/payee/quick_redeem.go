package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handleQuickRedeemEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "quick_redeem")

	var (
		ctokenAssetID string
		redeemTokens  decimal.Decimal
	)
	{
		var asset uuid.UUID
		_, e := mtg.Scan(body, &asset, &redeemTokens)
		if err := compound.Require(e == nil, "payee/mtgscan"); err != nil {
			log.WithError(err).Infoln("skip: scan memo failed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInvalidArgument)
		}

		redeemTokens = redeemTokens.Truncate(8)
		ctokenAssetID = asset.String()
		log = logger.FromContext(ctx).WithFields(logrus.Fields{
			"ctoken_asset_id": ctokenAssetID,
			"redeem_amount":   redeemTokens,
		})
		ctx = logger.WithContext(ctx, log)
	}

	market, err := w.mustGetMarketWithCToken(ctx, ctokenAssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketNotFound)
	}
	if err := compound.Require(!market.IsMarketClosed(), "payee/market-closed", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("market closed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrMarketClosed)
	}

	supply, err := w.mustGetSupply(ctx, userID, ctokenAssetID)
	if err != nil {
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrSupplyNotFound)
	}

	//accrue interest
	AccrueInterest(ctx, market, output.CreatedAt)

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Infoln("transactions.FindByTraceID")
		return err
	}

	underlyingAmount := redeemTokens.Mul(market.CurExchangeRate()).Truncate(8)
	if tx.ID == 0 {
		if err := compound.Require(
			redeemTokens.LessThanOrEqual(market.CTokens) &&
				market.RedeemAllowed(redeemTokens),
			"payee/redeem-disallowed",
		); err != nil {
			log.WithError(err).Infoln("skip: redeem not allowed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrRedeemNotAllowed)
		}

		if err := compound.Require(
			redeemTokens.LessThanOrEqual(supply.Collaterals),
			"payee/insufficient-collaterals",
		); err != nil {
			log.WithError(err).Infoln("skip: insufficient collaterals")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientCollaterals)
		}

		// check liqudity
		liquidity, err := w.accountService.CalculateAccountLiquidity(ctx, userID, market)
		if err != nil {
			log.WithError(err).Errorln("accountz.CalculateAccountLiquidity")
			return err
		}

		unpledgedTokenLiquidity := redeemTokens.Mul(market.CurExchangeRate()).Mul(market.CollateralFactor).Mul(market.Price)
		if unpledgedTokenLiquidity.GreaterThan(liquidity) {
			log.Errorf("insufficient liquidity, liquidity:%v, changed_liquidity:%v", liquidity, unpledgedTokenLiquidity)
			return w.handleRefundEventV0(ctx, output, userID, followID, core.ActionTypeQuickRedeem, core.ErrInsufficientLiquidity)
		}

		extra := core.NewTransactionExtra()
		extra.Put("asset_id", market.AssetID)
		extra.Put("amount", underlyingAmount)
		extra.Put("ctoken_asset_id", ctokenAssetID)
		extra.Put("ctokens", redeemTokens)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeQuickRedeem, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			log.WithError(err).Errorln("transactions.Create")
			return err
		}
	}

	// update supply
	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Sub(redeemTokens).Truncate(compound.MaxPricision)
		if err := w.supplyStore.Update(ctx, supply, output.ID); err != nil {
			log.WithError(err).Errorln("supply.Update")
			return err
		}
	}

	// transfer underlying asset
	if err := w.transferOut(
		ctx,
		userID,
		followID,
		output.TraceID,
		market.AssetID,
		underlyingAmount,
		&core.TransferAction{
			Source:   core.ActionTypeQuickRedeemTransfer,
			FollowID: followID,
		},
	); err != nil {
		return err
	}

	// update market
	if output.ID > market.Version {
		market.TotalCash = market.TotalCash.Sub(underlyingAmount).Truncate(compound.MaxPricision)
		market.CTokens = market.CTokens.Sub(redeemTokens).Truncate(compound.MaxPricision)
		if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("markets.Update")
			return err
		}
	}

	log.Infoln("quick redeem completed")
	return nil
}
