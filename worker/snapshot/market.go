package snapshot

import (
	"compound/core"
	"compound/internal/compound"
	"compound/pkg/id"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

var handleRequestMarketEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	if snapshot.AssetID != w.config.App.GasAssetID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	symbol := strings.ToUpper(action[core.ActionKeySymbol])

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		return e
	}

	// base info
	trace := id.UUIDFromString(fmt.Sprintf("market-base-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    w.config.App.GasAssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     core.GasCost,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceMarketResponse
		action[core.ActionKeySymbol] = symbol
		action[core.ActionKeyTotalCash] = market.TotalCash.String()
		action[core.ActionKeyTotalBorrows] = market.TotalBorrows.String()
		action[core.ActionKeyCTokens] = market.CTokens.String()
		action[core.ActionKeyPrice] = market.Price.String()
		action[core.ActionKeyBlock] = strconv.FormatInt(market.BlockNumber, 10)
		memoStr, e := action.Format()
		if e != nil {
			return e
		}
		input.Memo = memoStr
		if _, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin); e != nil {
			return e
		}
	}

	// rate info
	trace = id.UUIDFromString(fmt.Sprintf("market-rate-%s", snapshot.TraceID))
	input = mixin.TransferInput{
		AssetID:    w.config.App.GasAssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     core.GasCost,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		sRate, e := w.marketService.CurSupplyRate(ctx, market)
		if e != nil {
			return e
		}
		bRate, e := w.marketService.CurBorrowRate(ctx, market)
		if e != nil {
			return e
		}
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceMarketResponse
		action[core.ActionKeySymbol] = symbol
		action[core.ActionKeyUtilizationRate] = market.UtilizationRate.String()
		action[core.ActionKeyExchangeRate] = market.ExchangeRate.String()
		action[core.ActionKeySupplyRate] = sRate.Truncate(8).String()
		action[core.ActionKeyBorrowRate] = bRate.Truncate(8).String()
		memoStr, e := action.Format()
		if e != nil {
			return e
		}
		input.Memo = memoStr
		if _, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin); e != nil {
			return e
		}
	}

	return nil
}

var handleUpdateMarketEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	if snapshot.AssetID != w.config.App.GasAssetID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	if !w.config.IsAdmin(snapshot.OpponentID) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	symbol := strings.ToUpper(action[core.ActionKeySymbol])
	initialExchangeRate, e := decimal.NewFromString(action[core.ActionKeyInitExchangeRate])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if initialExchangeRate.LessThanOrEqual(decimal.Zero) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	reserveFactor, e := decimal.NewFromString(action[core.ActionKeyReserveFactor])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if reserveFactor.LessThanOrEqual(decimal.Zero) || reserveFactor.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	liquidationIncentive, e := decimal.NewFromString(action[core.ActionKeyLiquidationIncentive])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if liquidationIncentive.LessThan(compound.LiquidationIncentiveMin) || liquidationIncentive.GreaterThan(compound.LiquidationIncentiveMax) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	borrowCap, e := decimal.NewFromString(action[core.ActionKeyBorrowCap])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if borrowCap.LessThan(decimal.Zero) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	collateralFactor, e := decimal.NewFromString(action[core.ActionKeyCollateralFactor])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if collateralFactor.LessThan(decimal.Zero) || collateralFactor.GreaterThan(compound.CollateralFactorMax) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	closeFactor, e := decimal.NewFromString(action[core.ActionKeyCloseFactor])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if closeFactor.LessThan(compound.CloseFactorMin) || closeFactor.GreaterThan(compound.CloseFactorMax) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	baseRate, e := decimal.NewFromString(action[core.ActionKeyBaseRate])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if baseRate.LessThanOrEqual(decimal.Zero) || baseRate.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}

	multiplier, e := decimal.NewFromString(action[core.ActionKeyMultiPlier])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if multiplier.LessThanOrEqual(decimal.Zero) || multiplier.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	jumpMultiplier, e := decimal.NewFromString(action[core.ActionKeyJumpMultiPlier])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if jumpMultiplier.LessThanOrEqual(decimal.Zero) || jumpMultiplier.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	kink, e := decimal.NewFromString(action[core.ActionKeyKink])
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}
	if kink.LessThan(decimal.Zero) || jumpMultiplier.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrInvalidAmount)
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	if market.InitExchangeRate.GreaterThan(decimal.Zero) {
		if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
			return e
		}
	}

	market.InitExchangeRate = initialExchangeRate
	market.ReserveFactor = reserveFactor
	market.LiquidationIncentive = liquidationIncentive
	market.BorrowCap = borrowCap
	market.CollateralFactor = collateralFactor
	market.CloseFactor = closeFactor
	market.BaseRate = baseRate
	market.Multiplier = multiplier
	market.JumpMultiplier = jumpMultiplier
	market.Kink = kink

	if e = w.marketStore.Update(ctx, w.db, market); e != nil {
		return e
	}

	return nil
}

var handleAddMarketEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
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
		// market exists
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	} else {
		market := core.Market{
			Symbol:        symbol,
			AssetID:       assetID,
			CTokenAssetID: ctokenAssetID,
		}

		if e = w.marketStore.Save(ctx, w.db, &market); e != nil {
			return e
		}
	}

	return nil
}
