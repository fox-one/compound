package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// from user transfer
var handleSupplyEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "supply")
	supplyAmount := snapshot.Amount.Abs()
	userID := snapshot.OpponentID

	market, e := w.marketStore.Find(ctx, snapshot.AssetID)
	if e != nil {
		//refund to user
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	exchangeRate, e := w.marketService.CurExchangeRate(ctx, market)
	if e != nil {
		log.Errorln(e)
		return e
	}

	ctokens := snapshot.Amount.Div(exchangeRate).Truncate(8)
	if ctokens.LessThan(decimal.NewFromFloat(0.00000001)) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}

	return w.db.Tx(func(tx *db.DB) error {
		//update maket
		market.CTokens = market.CTokens.Add(ctokens)
		market.TotalCash = market.TotalCash.Add(supplyAmount)
		if e = w.marketStore.Update(ctx, tx, market); e != nil {
			log.Errorln(e)
			return e
		}

		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		//transfer ctoken to user
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceMint
		memoStr, e := memo.Format()
		if e != nil {
			log.Errorln(e)
			return e
		}
		trace := id.UUIDFromString(fmt.Sprintf("mint:%s", snapshot.TraceID))
		transfer := core.Transfer{
			AssetID:    market.CTokenAssetID,
			OpponentID: userID,
			Amount:     ctokens,
			TraceID:    trace,
			Memo:       memoStr,
		}

		if e = w.transferStore.Create(ctx, tx, &transfer); e != nil {
			log.Errorln(e)
			return e
		}

		return nil
	})
}

// from user, refund if error
var handlePledgeEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "pledge")
	ctokens := snapshot.Amount
	userID := snapshot.OpponentID
	ctokenAssetID := snapshot.AssetID

	market, e := w.marketStore.FindByCToken(ctx, ctokenAssetID)
	if e != nil {
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	if ctokens.GreaterThan(market.CTokens) {
		log.Errorln(errors.New("ctoken overflow"))
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrPledgeNotAllowed)
	}

	if market.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		log.Errorln(errors.New("pledge disallowed"))
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrPledgeNotAllowed)
	}

	return w.db.Tx(func(tx *db.DB) error {
		//accrue interest
		if e = w.marketService.AccrueInterest(ctx, tx, market, snapshot.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}

		supply, e := w.supplyStore.Find(ctx, userID, ctokenAssetID)
		if e != nil {
			if gorm.IsRecordNotFoundError(e) {
				//new
				supply = &core.Supply{
					UserID:        userID,
					CTokenAssetID: ctokenAssetID,
					Collaterals:   ctokens,
				}
				if e = w.supplyStore.Save(ctx, tx, supply); e != nil {
					log.Errorln(e)
					return e
				}
				return nil
			}
			log.Errorln(e)
			return e
		}
		//update supply
		supply.Collaterals = supply.Collaterals.Add(ctokens)
		e = w.supplyStore.Update(ctx, tx, supply)
		if e != nil {
			log.Errorln(e)
			return e
		}

		return nil
	})
}

// from system, transfer unpledged ctoken to user
var handleUnpledgeEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "unpledge")

	userID := snapshot.OpponentID
	symbol := strings.ToUpper(action[core.ActionKeySymbol])
	unpledgedTokens, e := decimal.NewFromString(action[core.ActionKeyCToken])
	if e != nil {
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
	if e != nil {
		log.Errorln(e)
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrSupplyNotFound)
	}

	//accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		log.Errorln(e)
		return e
	}

	if unpledgedTokens.GreaterThan(supply.Collaterals) {
		log.Errorln(errors.New("insufficient collaterals"))
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInsufficientCollaterals)
	}

	blockNum, e := w.blockService.GetBlock(ctx, snapshot.CreatedAt)
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
	unpledgedTokenLiquidity := unpledgedTokens.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
	if unpledgedTokenLiquidity.GreaterThan(liquidity) {
		log.Errorln(errors.New("insufficient liquidity"))
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInsufficientLiquidity)
	}

	return w.db.Tx(func(tx *db.DB) error {
		supply, e := w.supplyStore.Find(ctx, userID, market.CTokenAssetID)
		if e != nil {
			log.Errorln(e)
			return e
		}

		supply.Collaterals = supply.Collaterals.Sub(unpledgedTokens)
		if supply.Collaterals.LessThan(decimal.Zero) {
			supply.Collaterals = decimal.Zero
		}

		if e = w.supplyStore.Update(ctx, tx, supply); e != nil {
			log.Errorln(e)
			return e
		}

		//transfer ctoken to user
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceUnpledgeTransfer

		memoStr, e := action.Format()
		if e != nil {
			log.Errorln(e)
			return e
		}
		trace := id.UUIDFromString(fmt.Sprintf("unpledge-%s", snapshot.TraceID))
		transfer := core.Transfer{
			AssetID:    market.CTokenAssetID,
			OpponentID: userID,
			Amount:     unpledgedTokens,
			TraceID:    trace,
			Memo:       memoStr,
		}

		if e = w.transferStore.Create(ctx, tx, &transfer); e != nil {
			log.Errorln(e)
			return e
		}

		return nil
	})
}
