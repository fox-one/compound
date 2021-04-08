package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"compound/internal/compound"
	"context"
	"strings"

	"github.com/fox-one/pkg/logger"
	foxuuid "github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

func (w *Payee) handleUpdateMarketEvent(ctx context.Context, p *core.Proposal, req proposal.UpdateMarketReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithField("worker", "update-market")

	market, isRecordNotFound, e := w.marketStore.FindBySymbol(ctx, strings.ToUpper(req.Symbol))
	if e != nil {
		if isRecordNotFound {
			return nil
		}

		return e
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

	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	return nil
}

func (w *Payee) handleUpdateMarketAdvanceEvent(ctx context.Context, p *core.Proposal, req proposal.UpdateMarketAdvanceReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithField("worker", "update-market-advance")
	market, isRecordNotFound, e := w.marketStore.FindBySymbol(ctx, strings.ToUpper(req.Symbol))
	if e != nil {
		if isRecordNotFound {
			return nil
		}

		return e
	}

	if e = w.marketService.AccrueInterest(ctx, market, output.CreatedAt); e != nil {
		return e
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

	if e = w.marketStore.Update(ctx, market, output.ID); e != nil {
		log.Errorln(e)
		return e
	}

	// market transaction
	marketTransaction := core.BuildMarketUpdateTransaction(ctx, market, foxuuid.Modify(output.TraceID, "update_market"))
	if e = w.transactionStore.Create(ctx, marketTransaction); e != nil {
		log.WithError(e).Errorln("create transaction error")
		return e
	}

	return nil
}
