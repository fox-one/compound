package compound

import (
	"compound/core"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// CurBorrowRate current borrow APY
func CurBorrowRate(market *core.Market) decimal.Decimal {
	borrowRatePerBlock := curBorrowRatePerBlockInternal(market)
	return borrowRatePerBlock.Mul(BlocksPerYear).Truncate(MaxPricision)
}

//
func curBorrowRatePerBlockInternal(market *core.Market) decimal.Decimal {
	return GetBorrowRatePerBlock(
		UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves),
		market.BaseRate,
		market.Multiplier,
		market.JumpMultiplier,
		market.Kink,
	)
}

// CurSupplyRate current supply APY
func CurSupplyRate(market *core.Market) decimal.Decimal {
	supplyRatePerBlock := curSupplyRatePerBlockInternal(market)
	return supplyRatePerBlock.Mul(BlocksPerYear).Truncate(MaxPricision)
}

//
func curSupplyRatePerBlockInternal(market *core.Market) decimal.Decimal {
	return GetSupplyRatePerBlock(
		UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves),
		market.BaseRate,
		market.Multiplier,
		market.JumpMultiplier,
		market.Kink,
		market.ReserveFactor,
	)
}

// AccrueInterest accrue interest market per block(15 seconds)
//
// Accruing interest only occurs when there is a behavior that causes changes in market transaction data, such as supply, borrow, pledge, unpledge, redeem, repay, price updating
func AccrueInterest(ctx context.Context, market *core.Market, time time.Time) error {
	blockNum, err := GetBlockByTime(ctx, time)
	if err != nil {
		return err
	}

	if !market.BorrowIndex.IsPositive() {
		market.BorrowIndex = decimal.New(1, 0)
	}

	if blockDelta := blockNum - market.BlockNumber; blockDelta > 0 {
		borrowRate := curBorrowRatePerBlockInternal(market)
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
	uRate := UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves)
	//exchange rate
	exchangeRate := GetExchangeRate(market.TotalCash, market.TotalBorrows, market.Reserves, market.CTokens, market.InitExchangeRate)
	supplyRate := curSupplyRatePerBlockInternal(market)
	borrowRate := curBorrowRatePerBlockInternal(market)

	market.UtilizationRate = uRate.Truncate(16)
	market.ExchangeRate = exchangeRate.Truncate(16)
	market.SupplyRatePerBlock = supplyRate.Truncate(16)
	market.BorrowRatePerBlock = borrowRate.Truncate(16)

	return nil
}
