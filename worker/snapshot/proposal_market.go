package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"compound/internal/compound"
	"context"
	"strings"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithField("worker", "add-market")

	log.Infof("asset:%s", req.AssetID)
	market, e := w.marketStore.Find(ctx, req.AssetID)
	if e != nil {
		return e
	}
	if market.ID == 0 {
		market = &core.Market{
			Symbol:               strings.ToUpper(req.Symbol),
			AssetID:              req.AssetID,
			CTokenAssetID:        req.CTokenAssetID,
			InitExchangeRate:     req.InitExchange,
			ReserveFactor:        req.ReserveFactor,
			LiquidationIncentive: req.LiquidationIncentive,
			BorrowCap:            req.BorrowCap,
			BorrowIndex:          decimal.New(1, 0),
			CollateralFactor:     req.CollateralFactor,
			CloseFactor:          req.CloseFactor,
			BaseRate:             req.BaseRate,
			Multiplier:           req.Multiplier,
			JumpMultiplier:       req.JumpMultiplier,
			Kink:                 req.Kink,
			Price:                req.Price,
			PriceThreshold:       req.PriceThreshold,
			PriceUpdatedAt:       output.CreatedAt,
			Status:               core.MarketStatusClose,
		}

		if e = w.marketStore.Save(ctx, market); e != nil {
			return e
		}

		return nil
	}

	if market.InitExchangeRate.GreaterThan(decimal.Zero) {
		if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
			return e
		}
	}

	if req.InitExchange.GreaterThan(decimal.Zero) {
		market.InitExchangeRate = req.InitExchange
	}

	if req.ReserveFactor.GreaterThan(decimal.Zero) && req.ReserveFactor.LessThan(decimal.NewFromInt(1)) {
		market.ReserveFactor = req.ReserveFactor
	}

	if req.LiquidationIncentive.GreaterThanOrEqual(compound.LiquidationIncentiveMin) && req.LiquidationIncentive.LessThanOrEqual(compound.LiquidationIncentiveMax) {
		market.LiquidationIncentive = req.LiquidationIncentive
	}

	if req.CollateralFactor.GreaterThanOrEqual(decimal.Zero) && req.CollateralFactor.LessThanOrEqual(compound.CollateralFactorMax) {
		market.CollateralFactor = req.CollateralFactor
	}

	if req.BaseRate.GreaterThan(decimal.Zero) && req.BaseRate.LessThan(decimal.NewFromInt(1)) {
		market.BaseRate = req.BaseRate
	}

	if req.BorrowCap.GreaterThanOrEqual(decimal.Zero) {
		market.BorrowCap = req.BorrowCap
	}

	if req.CloseFactor.GreaterThanOrEqual(compound.CloseFactorMin) && req.CloseFactor.LessThanOrEqual(compound.CloseFactorMax) {
		market.CloseFactor = req.CloseFactor
	}

	if req.Multiplier.GreaterThan(decimal.Zero) && req.Multiplier.LessThan(decimal.NewFromInt(1)) {
		market.Multiplier = req.Multiplier
	}

	if req.JumpMultiplier.GreaterThanOrEqual(decimal.Zero) && req.JumpMultiplier.LessThan(decimal.NewFromInt(1)) {
		market.JumpMultiplier = req.JumpMultiplier
	}

	if req.Kink.GreaterThanOrEqual(decimal.Zero) && req.Kink.LessThan(decimal.NewFromInt(1)) {
		market.Kink = req.Kink
	}

	if req.PriceThreshold > 0 {
		market.PriceThreshold = req.PriceThreshold
	} else if req.PriceThreshold < 0 {
		market.PriceThreshold = 0
	}

	if req.Price.IsPositive() {
		market.Price = req.Price
	}

	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	return nil
}
