package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/compound"
	"context"
	"strings"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handleMarketEvent(ctx context.Context, p *core.Proposal, req proposal.MarketReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "upsert-market",
		"asset":    req.AssetID,
	})

	market, err := w.marketStore.Find(ctx, req.AssetID)
	if err != nil {
		log.WithError(err).Errorln("markets.Find")
		return err
	}

	if market.ID == 0 {
		market = &core.Market{
			Symbol:               strings.ToUpper(req.Symbol),
			AssetID:              req.AssetID,
			CTokenAssetID:        req.CTokenAssetID,
			InitExchangeRate:     req.InitExchange,
			ExchangeRate:         req.InitExchange,
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
			Version:              output.ID,
		}

		if err := w.marketStore.Create(ctx, market); err != nil {
			log.WithError(err).Errorln("markets.Create")
			return err
		}
		log.Infoln("market created")
		return nil
	}

	if market.Version >= output.ID {
		return nil
	}

	one := decimal.New(1, 0)

	AccrueInterest(ctx, market, output.CreatedAt)

	if req.InitExchange.IsPositive() {
		market.InitExchangeRate = req.InitExchange
	}

	if req.ReserveFactor.IsPositive() && req.ReserveFactor.LessThan(one) {
		market.ReserveFactor = req.ReserveFactor
	}

	if req.LiquidationIncentive.GreaterThanOrEqual(compound.LiquidationIncentiveMin) &&
		req.LiquidationIncentive.LessThanOrEqual(compound.LiquidationIncentiveMax) {
		market.LiquidationIncentive = req.LiquidationIncentive
	}

	if !req.CollateralFactor.IsNegative() &&
		req.CollateralFactor.LessThanOrEqual(compound.CollateralFactorMax) {
		market.CollateralFactor = req.CollateralFactor
	}

	if req.BaseRate.IsPositive() && req.BaseRate.LessThan(one) {
		market.BaseRate = req.BaseRate
	}

	if !req.BorrowCap.IsNegative() {
		market.BorrowCap = req.BorrowCap
	}

	if req.CloseFactor.GreaterThanOrEqual(compound.CloseFactorMin) &&
		req.CloseFactor.LessThanOrEqual(compound.CloseFactorMax) {
		market.CloseFactor = req.CloseFactor
	}

	if req.Multiplier.IsPositive() && req.Multiplier.LessThan(one) {
		market.Multiplier = req.Multiplier
	}

	if !req.JumpMultiplier.IsNegative() && req.JumpMultiplier.LessThan(one) {
		market.JumpMultiplier = req.JumpMultiplier
	}

	if !req.Kink.IsNegative() && req.Kink.LessThan(one) {
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

	if err := w.marketStore.Update(ctx, market, output.ID); err != nil {
		log.WithError(err).Errorln("markets.Update")
		return err
	}
	log.Infoln("market updated")
	return nil
}
