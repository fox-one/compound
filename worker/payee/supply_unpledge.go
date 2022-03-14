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

// handle unpledge event
func (w *Payee) handleUnpledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "unpledge")

	var (
		ctokenAssetID   string
		unpledgedAmount decimal.Decimal
	)
	{
		var asset uuid.UUID
		_, e := mtg.Scan(body, &asset, &unpledgedAmount)
		if err := compound.Require(e == nil, "payee/mtgscan"); err != nil {
			log.WithError(err).Infoln("skip: scan memo failed")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeUnpledge, core.ErrInvalidArgument)
		}

		unpledgedAmount = unpledgedAmount.Truncate(8)
		ctokenAssetID = asset.String()
		log = logger.FromContext(ctx).WithFields(logrus.Fields{
			"ctoken_asset_id": ctokenAssetID,
			"unpledge_amount": unpledgedAmount,
		})
		ctx = logger.WithContext(ctx, log)
	}

	market, err := w.mustGetMarketWithCToken(ctx, ctokenAssetID)
	if err != nil {
		return err
	}

	if market.Version >= output.ID {
		log.Infoln("skip: output.ID outdated")
		return nil
	}

	if err := compound.Require(!market.IsMarketClosed(), "payee/market-closed", compound.FlagRefund); err != nil {
		log.WithError(err).Infoln("market closed")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeUnpledge, core.ErrMarketClosed)
	}

	//accrue interest
	AccrueInterest(ctx, market, output.CreatedAt)

	supply, err := w.mustGetSupply(ctx, userID, ctokenAssetID)
	if err != nil {
		log.WithError(err).Infoln("invalid supply")
		return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeUnpledge, core.ErrSupplyNotFound)
	}

	tx, err := w.transactionStore.FindByTraceID(ctx, output.TraceID)
	if err != nil {
		log.WithError(err).Errorln("transactions.FindByTraceID")
		return err
	}

	if tx.ID == 0 {
		if err := compound.Require(
			unpledgedAmount.LessThan(supply.Collaterals),
			"payee/insufficient-collaterals",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Infoln("refund")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeUnpledge, core.ErrInsufficientCollaterals)
		}

		// check liqudity
		liquidity, err := w.accountService.CalculateAccountLiquidity(ctx, userID, market)
		if err != nil {
			log.WithError(err).Errorln("accountz.CalculateAccountLiquidity")
			return err
		}

		if err := compound.Require(
			unpledgedAmount.Mul(market.ExchangeRate).Mul(market.CollateralFactor).Mul(market.Price).LessThanOrEqual(liquidity),
			"payee/insufficient-borrow-balance",
			compound.FlagRefund,
		); err != nil {
			log.WithError(err).Infoln("refund")
			return w.returnOrRefundError(ctx, err, output, userID, followID, core.ActionTypeUnpledge, core.ErrInsufficientLiquidity)
		}

		extra := core.NewTransactionExtra()
		extra.Put("ctoken_asset_id", ctokenAssetID)
		extra.Put("amount", unpledgedAmount)

		tx = core.BuildTransactionFromOutput(ctx, userID, followID, core.ActionTypeUnpledge, output, extra)
		if err := w.transactionStore.Create(ctx, tx); err != nil {
			log.WithError(err).Errorln("transactions.Create")
			return err
		}
	}

	if output.ID > supply.Version {
		supply.Collaterals = supply.Collaterals.Sub(unpledgedAmount).Truncate(compound.MaxPricision)
		if err := w.supplyStore.Update(ctx, supply, output.ID); err != nil {
			log.WithError(err).Errorln("supplies.Update")
			return err
		}
	}

	// add transfer
	if err := w.transferOut(
		ctx,
		userID,
		followID,
		output.TraceID,
		market.CTokenAssetID,
		unpledgedAmount,
		&core.TransferAction{
			Source:   core.ActionTypeUnpledgeTransfer,
			FollowID: followID,
		},
	); err != nil {
		return err
	}

	if output.ID > market.Version {
		if err = w.marketStore.Update(ctx, market, output.ID); err != nil {
			log.WithError(err).Errorln("update market error")
			return err
		}
	}

	log.Infoln("unpledge completed")
	return nil
}
