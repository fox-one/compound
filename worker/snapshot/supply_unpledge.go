package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleUnpledgeEvent(ctx context.Context, output *core.Output, userID, followID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "unpledge")

	var ctokenAsset uuid.UUID
	var unpledgedAmount decimal.Decimal

	if _, err := mtg.Scan(body, &ctokenAsset, &unpledgedAmount); err != nil {
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrInvalidArgument, "")
	}

	ctokenAssetID := ctokenAsset.String()
	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrMarketNotFound, "")
	}

	supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		log.Errorln(e)
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrSupplyNotFound, "")
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, output.UpdatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	if unpledgedAmount.GreaterThan(supply.Collaterals) {
		log.Errorln(errors.New("insufficient collaterals"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrInsufficientCollaterals, "")
	}

	blockNum, e := w.blockService.GetBlock(ctx, output.UpdatedAt)
	if e != nil {
		log.Errorln(e)
		return e
	}

	// check liqudity
	liquidity, e := w.accountService.CalculateAccountLiquidity(ctx, userID, blockNum)
	if e != nil {
		log.Errorln(e)
		return e
	}

	price, e := w.priceService.GetCurrentUnderlyingPrice(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}

	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}
	unpledgedTokenLiquidity := unpledgedAmount.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
	if unpledgedTokenLiquidity.GreaterThan(liquidity) {
		log.Errorln(errors.New("insufficient liquidity"))
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrInsufficientLiquidity, "")
	}

	return w.db.Tx(func(tx *db.DB) error {
		supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
		if e != nil {
			log.Errorln(e)
			return e
		}

		supply.Collaterals = supply.Collaterals.Sub(unpledgedAmount).Truncate(16)
		if supply.Collaterals.LessThan(decimal.Zero) {
			supply.Collaterals = decimal.Zero
		}

		if e = w.supplyStore.Update(ctx, tx, supply); e != nil {
			log.Errorln(e)
			return e
		}

		transferAction := core.TransferAction{
			Source:        core.ActionTypeUnpledgeTransfer,
			TransactionID: followID,
		}

		return w.transferOut(ctx, userID, followID, output.TraceID, market.CTokenAssetID, unpledgedAmount, &transferAction)
	})
}
