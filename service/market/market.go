package market

import (
	"compound/core"
	"compound/internal/compound"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type service struct {
	blockSrv core.IBlockService
}

// New new market service
func New(
	blockSrv core.IBlockService,
) core.IMarketService {
	return &service{
		blockSrv: blockSrv,
	}
}

func (s *service) CurUtilizationRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	ur := market.UtilizationRate
	if !ur.IsPositive() {
		return decimal.Zero, nil
	}

	return ur, nil
}
func (s *service) CurExchangeRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	er := market.ExchangeRate
	if !er.IsPositive() {
		return market.InitExchangeRate, nil
	}

	return er, nil
}
func (s *service) CurBorrowRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	br := market.BorrowRatePerBlock
	if !br.IsPositive() {
		return s.curBorrowRatePerBlockInternal(ctx, market)
	}

	return br, nil
}
func (s *service) CurSupplyRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	sr := market.SupplyRatePerBlock
	if !sr.IsPositive() {
		return s.curSupplyRatePerBlockInternal(ctx, market)
	}

	return sr, nil
}

//
func (s *service) curUtilizationRateInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	rate := compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves)
	return rate, nil
}

func (s *service) curExchangeRateInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	if !market.CTokens.IsPositive() {
		return market.InitExchangeRate, nil
	}

	rate := compound.GetExchangeRate(market.TotalCash, market.TotalBorrows, market.Reserves, market.CTokens, market.InitExchangeRate)

	return rate, nil
}

// CurBorrowRate current borrow APY
func (s *service) CurBorrowRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	borrowRatePerBlock, e := s.curBorrowRatePerBlockInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return borrowRatePerBlock.Mul(compound.BlocksPerYear).Truncate(compound.MaxPricision), nil
}

//
func (s *service) curBorrowRatePerBlockInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.curUtilizationRateInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetBorrowRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink)

	return rate, nil
}

// CurSupplyRate current supply APY
func (s *service) CurSupplyRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	supplyRatePerBlock, e := s.curSupplyRatePerBlockInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return supplyRatePerBlock.Mul(compound.BlocksPerYear).Truncate(compound.MaxPricision), nil
}

//
func (s *service) curSupplyRatePerBlockInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.curUtilizationRateInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetSupplyRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink, market.ReserveFactor)

	return rate, nil
}

//
func (s *service) CurTotalBorrows(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.TotalBorrows, nil
}

//
func (s *service) CurTotalReserves(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.Reserves, nil
}

// AccrueInterest accrue interest market per block(15 seconds)
//
// Accruing interest only occurs when there is a behavior that causes changes in market transaction data, such as supply, borrow, pledge, unpledge, redeem, repay, price updating
func (s *service) AccrueInterest(ctx context.Context, market *core.Market, time time.Time) error {
	blockNumberPrior := market.BlockNumber

	blockNum, e := s.blockSrv.GetBlock(ctx, time)
	if e != nil {
		return e
	}

	if !market.BorrowIndex.IsPositive() {
		market.BorrowIndex = decimal.New(1, 0)
	}

	blockDelta := blockNum - blockNumberPrior
	if blockDelta > 0 {
		borrowRate, e := s.curBorrowRatePerBlockInternal(ctx, market)
		if e != nil {
			return e
		}

		timesBorrowRate := borrowRate.Mul(decimal.NewFromInt(blockDelta))
		interestAccumulated := market.TotalBorrows.Mul(timesBorrowRate).Truncate(compound.MaxPricision)

		market.BlockNumber = blockNum
		market.TotalBorrows = market.TotalBorrows.Add(interestAccumulated)
		market.Reserves = market.Reserves.Add(interestAccumulated.Mul(market.ReserveFactor).Truncate(compound.MaxPricision))
		market.BorrowIndex = market.BorrowIndex.Add(
			timesBorrowRate.Mul(market.BorrowIndex).
				Shift(compound.MaxPricision).Ceil().Shift(-compound.MaxPricision))
	}

	//utilization rate
	uRate, e := s.curUtilizationRateInternal(ctx, market)
	if e != nil {
		return e
	}

	//exchange rate
	exchangeRate, e := s.curExchangeRateInternal(ctx, market)
	if e != nil {
		return e
	}

	supplyRate, e := s.curSupplyRatePerBlockInternal(ctx, market)
	if e != nil {
		return e
	}

	borrowRate, e := s.curBorrowRatePerBlockInternal(ctx, market)
	if e != nil {
		return e
	}

	market.UtilizationRate = uRate.Truncate(16)
	market.ExchangeRate = exchangeRate.Truncate(16)
	market.SupplyRatePerBlock = supplyRate.Truncate(16)
	market.BorrowRatePerBlock = borrowRate.Truncate(16)

	return nil
}
