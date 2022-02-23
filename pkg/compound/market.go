package compound

import (
	"compound/core"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

func CurUtilizationRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	ur := market.UtilizationRate
	if !ur.IsPositive() {
		return decimal.Zero, nil
	}

	return ur, nil
}
func CurExchangeRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	er := market.ExchangeRate
	if !er.IsPositive() {
		return market.InitExchangeRate, nil
	}

	return er, nil
}
func CurBorrowRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	br := market.BorrowRatePerBlock
	if !br.IsPositive() {
		return curBorrowRatePerBlockInternal(ctx, market)
	}

	return br, nil
}
func CurSupplyRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	sr := market.SupplyRatePerBlock
	if !sr.IsPositive() {
		return curSupplyRatePerBlockInternal(ctx, market)
	}

	return sr, nil
}

//
func curUtilizationRateInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	rate := UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves)
	return rate, nil
}

func curExchangeRateInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	if !market.CTokens.IsPositive() {
		return market.InitExchangeRate, nil
	}

	rate := GetExchangeRate(market.TotalCash, market.TotalBorrows, market.Reserves, market.CTokens, market.InitExchangeRate)

	return rate, nil
}

// CurBorrowRate current borrow APY
func CurBorrowRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	borrowRatePerBlock, e := curBorrowRatePerBlockInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return borrowRatePerBlock.Mul(BlocksPerYear).Truncate(MaxPricision), nil
}

//
func curBorrowRatePerBlockInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := curUtilizationRateInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := GetBorrowRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink)

	return rate, nil
}

// CurSupplyRate current supply APY
func CurSupplyRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	supplyRatePerBlock, e := curSupplyRatePerBlockInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return supplyRatePerBlock.Mul(BlocksPerYear).Truncate(MaxPricision), nil
}

//
func curSupplyRatePerBlockInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := curUtilizationRateInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := GetSupplyRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink, market.ReserveFactor)

	return rate, nil
}

//
func CurTotalBorrows(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.TotalBorrows, nil
}

//
func CurTotalReserves(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.Reserves, nil
}

// AccrueInterest accrue interest market per block(15 seconds)
//
// Accruing interest only occurs when there is a behavior that causes changes in market transaction data, such as supply, borrow, pledge, unpledge, redeem, repay, price updating
func AccrueInterest(ctx context.Context, market *core.Market, time time.Time) error {
	blockNumberPrior := market.BlockNumber

	blockNum, e := GetBlockByTime(ctx, time)
	if e != nil {
		return e
	}

	if !market.BorrowIndex.IsPositive() {
		market.BorrowIndex = decimal.New(1, 0)
	}

	blockDelta := blockNum - blockNumberPrior
	if blockDelta > 0 {
		borrowRate, e := curBorrowRatePerBlockInternal(ctx, market)
		if e != nil {
			return e
		}

		timesBorrowRate := borrowRate.Mul(decimal.NewFromInt(blockDelta))
		interestAccumulated := market.TotalBorrows.Mul(timesBorrowRate).Truncate(MaxPricision)

		market.BlockNumber = blockNum
		market.TotalBorrows = market.TotalBorrows.Add(interestAccumulated)
		market.Reserves = market.Reserves.Add(interestAccumulated.Mul(market.ReserveFactor).Truncate(MaxPricision))
		market.BorrowIndex = market.BorrowIndex.Add(
			timesBorrowRate.Mul(market.BorrowIndex).
				Shift(MaxPricision).Ceil().Shift(-MaxPricision))
	}

	//utilization rate
	uRate, e := curUtilizationRateInternal(ctx, market)
	if e != nil {
		return e
	}

	//exchange rate
	exchangeRate, e := curExchangeRateInternal(ctx, market)
	if e != nil {
		return e
	}

	supplyRate, e := curSupplyRatePerBlockInternal(ctx, market)
	if e != nil {
		return e
	}

	borrowRate, e := curBorrowRatePerBlockInternal(ctx, market)
	if e != nil {
		return e
	}

	market.UtilizationRate = uRate.Truncate(16)
	market.ExchangeRate = exchangeRate.Truncate(16)
	market.SupplyRatePerBlock = supplyRate.Truncate(16)
	market.BorrowRatePerBlock = borrowRate.Truncate(16)

	return nil
}
