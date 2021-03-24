package compound

import (
	"github.com/shopspring/decimal"
)

var (
	// SecondsPerBlock seconds per block
	SecondsPerBlock int64 = 15
	// BlocksPerYear blocks per year
	BlocksPerYear = decimal.NewFromInt(2102400)
	// CloseFactorMin min of close factor, must be strictly greater than this value
	CloseFactorMin = decimal.NewFromFloat(0.05)
	// CloseFactorMax max of close factor, must not exceed this value
	CloseFactorMax = decimal.NewFromFloat(0.9)
	// CollateralFactorMax max of collateral factor, may exceed this value [0, 0.9]
	CollateralFactorMax = decimal.NewFromFloat(0.9)
	// LiquidationIncentiveMin must be no less than this value
	LiquidationIncentiveMin = decimal.NewFromFloat(0.01)
	// LiquidationIncentiveMax must be no greater than this value
	LiquidationIncentiveMax = decimal.NewFromFloat(0.9)
	// MaxPricision max pricision
	MaxPricision int32 = 16
)

// UtilizationRate utilization rate
// utilization_rate = market.total_borrows/(market.total_cash + market.borrows - market.reserves)
func UtilizationRate(cash, borrows, reserves decimal.Decimal) decimal.Decimal {
	total := cash.Add(borrows).Sub(reserves)
	if total.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	return borrows.Div(cash.Add(borrows).Sub(reserves)).Truncate(MaxPricision)
}

// GetExchangeRate exchange rate
// exchange_rate = (market.total_cash + market.total_borrows - market.reserves) / market.tocken_supply
func GetExchangeRate(totalCash, totalBorrows, totalReserves, tokenSupply, initialExchangeRate decimal.Decimal) decimal.Decimal {
	if tokenSupply.Equal(decimal.Zero) {
		return initialExchangeRate
	}

	return totalCash.Add(totalBorrows).Sub(totalReserves).Div(tokenSupply).Truncate(MaxPricision)
}

// GetBorrowRatePerBlock borrowRate per block
func GetBorrowRatePerBlock(utilizationRate, baseRate, multiplier, jumpMultiplier, kink decimal.Decimal) decimal.Decimal {
	if kink.Equal(decimal.Zero) ||
		utilizationRate.LessThanOrEqual(kink) {
		return utilizationRate.Mul(GetMultiplierPerBlock(multiplier)).Add(GetBaseRatePerBlock(baseRate)).Truncate(MaxPricision)
	}

	normalRate := kink.Mul(GetMultiplierPerBlock(multiplier)).Add(GetBaseRatePerBlock(baseRate))
	excessUtilRate := utilizationRate.Sub(kink)
	return excessUtilRate.Mul(GetJumpMultiplierPerBlock(jumpMultiplier)).Add(normalRate).Truncate(MaxPricision)
}

// GetSupplyRatePerBlock supply rate per block
func GetSupplyRatePerBlock(utilizationRate, baseRate, multiplier, jumpMultiplier, kink, reserveFactor decimal.Decimal) decimal.Decimal {
	borrowRate := GetBorrowRatePerBlock(utilizationRate, baseRate, multiplier, jumpMultiplier, kink)
	oneMinusReserveFactor := decimal.NewFromInt(1).Sub(reserveFactor)
	rateToPool := borrowRate.Mul(oneMinusReserveFactor)
	return utilizationRate.Mul(rateToPool).Truncate(MaxPricision)
}

// GetBaseRatePerBlock base rate per block
func GetBaseRatePerBlock(baseRate decimal.Decimal) decimal.Decimal {
	return baseRate.Div(BlocksPerYear).Truncate(MaxPricision)
}

// GetMultiplierPerBlock multiplier per block
func GetMultiplierPerBlock(multiplier decimal.Decimal) decimal.Decimal {
	return multiplier.Div(BlocksPerYear).Truncate(MaxPricision)
}

// GetJumpMultiplierPerBlock jump multiplier per block
func GetJumpMultiplierPerBlock(jumpMultiplier decimal.Decimal) decimal.Decimal {
	return jumpMultiplier.Div(BlocksPerYear).Truncate(MaxPricision)
}
