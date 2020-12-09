package snapshot

import (
	"compound/core"
	"compound/internal/compound"
	"context"
	"strings"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

var handleUpdateMarketEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "update-market")
	if snapshot.AssetID != w.config.App.GasAssetID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	if !w.config.IsAdmin(snapshot.OpponentID) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	symbol := strings.ToUpper(action[core.ActionKeySymbol])

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	if market.InitExchangeRate.GreaterThan(decimal.Zero) {
		if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
			log.Errorln(e)
			return e
		}
	}

	initialExchangeRate, e := decimal.NewFromString(action[core.ActionKeyInitExchangeRate])
	if e == nil && initialExchangeRate.GreaterThan(decimal.Zero) {
		market.InitExchangeRate = initialExchangeRate
	}

	reserveFactor, e := decimal.NewFromString(action[core.ActionKeyReserveFactor])
	if e == nil && reserveFactor.GreaterThan(decimal.Zero) && reserveFactor.LessThan(decimal.NewFromInt(1)) {
		market.ReserveFactor = reserveFactor
	}

	liquidationIncentive, e := decimal.NewFromString(action[core.ActionKeyLiquidationIncentive])
	if e == nil && liquidationIncentive.GreaterThanOrEqual(compound.LiquidationIncentiveMin) && liquidationIncentive.LessThanOrEqual(compound.LiquidationIncentiveMax) {
		market.LiquidationIncentive = liquidationIncentive
	}

	borrowCap, e := decimal.NewFromString(action[core.ActionKeyBorrowCap])
	if e == nil && borrowCap.GreaterThanOrEqual(decimal.Zero) {
		market.BorrowCap = borrowCap
	}

	collateralFactor, e := decimal.NewFromString(action[core.ActionKeyCollateralFactor])
	if e == nil && collateralFactor.GreaterThanOrEqual(decimal.Zero) && collateralFactor.LessThanOrEqual(compound.CollateralFactorMax) {
		market.CollateralFactor = collateralFactor
	}

	closeFactor, e := decimal.NewFromString(action[core.ActionKeyCloseFactor])
	if e == nil && closeFactor.GreaterThanOrEqual(compound.CloseFactorMin) && closeFactor.LessThanOrEqual(compound.CloseFactorMax) {
		market.CloseFactor = closeFactor
	}

	baseRate, e := decimal.NewFromString(action[core.ActionKeyBaseRate])
	if e == nil && baseRate.GreaterThan(decimal.Zero) && baseRate.LessThan(decimal.NewFromInt(1)) {
		market.BaseRate = baseRate
	}

	multiplier, e := decimal.NewFromString(action[core.ActionKeyMultiPlier])
	if e == nil && multiplier.GreaterThan(decimal.Zero) && multiplier.LessThan(decimal.NewFromInt(1)) {
		market.Multiplier = multiplier
	}

	jumpMultiplier, e := decimal.NewFromString(action[core.ActionKeyJumpMultiPlier])
	if e == nil && jumpMultiplier.GreaterThanOrEqual(decimal.Zero) && jumpMultiplier.LessThan(decimal.NewFromInt(1)) {
		market.JumpMultiplier = jumpMultiplier
	}

	kink, e := decimal.NewFromString(action[core.ActionKeyKink])
	if e == nil && kink.GreaterThanOrEqual(decimal.Zero) && kink.LessThan(decimal.NewFromInt(1)) {
		market.Kink = kink
	}

	if e = w.marketStore.Update(ctx, w.db, market); e != nil {
		log.Errorln(e)
		return e
	}

	return nil
}

var handleAddMarketEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "add-market")
	if snapshot.AssetID != w.config.App.GasAssetID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	if !w.config.IsAdmin(snapshot.OpponentID) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	symbol := strings.ToUpper(action[core.ActionKeySymbol])
	assetID := action[core.ActionKeyAssetID]
	ctokenAssetID := action[core.ActionKeyCTokenAssetID]

	_, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e == nil {
		log.Errorln(e)
		// market exists
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	market := core.Market{
		Symbol:        symbol,
		AssetID:       assetID,
		CTokenAssetID: ctokenAssetID,
	}

	if e = w.marketStore.Save(ctx, w.db, &market); e != nil {
		log.Errorln(e)
		return e
	}

	return nil
}
