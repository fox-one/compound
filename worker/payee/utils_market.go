package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// AccrueInterest accrue interest market per block(15 seconds)
//
// Accruing interest only occurs when there is a behavior that causes changes in market transaction data, such as supply, borrow, pledge, unpledge, redeem, repay, price updating
func AccrueInterest(ctx context.Context, market *core.Market, time time.Time) {
	blockNum, err := compound.GetBlockByTime(ctx, time)
	if err != nil {
		panic(err)
	}

	if !market.BorrowIndex.IsPositive() {
		market.BorrowIndex = decimal.New(1, 0)
	}

	if blockDelta := blockNum - market.BlockNumber; blockDelta > 0 {
		borrowRate := curBorrowRatePerBlockInternal(market)
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
	uRate := compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves)
	//exchange rate
	exchangeRate := compound.GetExchangeRate(market.TotalCash, market.TotalBorrows, market.Reserves, market.CTokens, market.InitExchangeRate)
	supplyRate := curSupplyRatePerBlockInternal(market)
	borrowRate := curBorrowRatePerBlockInternal(market)

	market.UtilizationRate = uRate.Truncate(16)
	market.ExchangeRate = exchangeRate.Truncate(16)
	market.SupplyRatePerBlock = supplyRate.Truncate(16)
	market.BorrowRatePerBlock = borrowRate.Truncate(16)
}

//
func curBorrowRatePerBlockInternal(market *core.Market) decimal.Decimal {
	return compound.GetBorrowRatePerBlock(
		compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves),
		market.BaseRate,
		market.Multiplier,
		market.JumpMultiplier,
		market.Kink,
	)
}

//
func curSupplyRatePerBlockInternal(market *core.Market) decimal.Decimal {
	return compound.GetSupplyRatePerBlock(
		compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves),
		market.BaseRate,
		market.Multiplier,
		market.JumpMultiplier,
		market.Kink,
		market.ReserveFactor,
	)
}
