package compound

import (
	"errors"

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
	LiquidationIncentiveMin = decimal.NewFromFloat(1.0)
	// LiquidationIncentiveMax must be no greater than this value
	LiquidationIncentiveMax = decimal.NewFromFloat(1.5)
	// InitialExchangeRate initial exchange rate
	InitialExchangeRate = decimal.NewFromInt(1)
)

var (
	// ErrUnsupported unsupported error
	ErrUnsupported = errors.New("unsupported")
)

// GetBorrowRate get borrow rate
func GetBorrowRate(cash, borrows, reserves decimal.Decimal) decimal.Decimal {
	return decimal.Zero
}

// GetSupplyRate get supply rate
func GetSupplyRate(cash, borrows, reserves, reserveFactor decimal.Decimal) decimal.Decimal {
	return decimal.Zero
}

// UtilizationRate utilization rate
func UtilizationRate(cash, borrows, reserves decimal.Decimal) decimal.Decimal {
	return borrows.Div(cash.Add(borrows).Sub(reserves))
}

// ReservesNew calculate new reserves
func ReservesNew(interestAccumulated, reserveFactor decimal.Decimal) decimal.Decimal {
	return interestAccumulated.Mul(reserveFactor)
}

// GetExchangeRate exchange rate
func GetExchangeRate(totalCash, totalBorrows, totalReserves, tokenSupply, initialExchangeRate decimal.Decimal) decimal.Decimal {
	if tokenSupply.Equal(decimal.Zero) {
		return initialExchangeRate
	}

	return totalCash.Add(totalBorrows).Sub(totalReserves).Div(tokenSupply)
}

// GetBorrowRatePerBlock borrowRate per block
func GetBorrowRatePerBlock(utilizationRate, baseRate, multiplier, jumpMultiplier, kink decimal.Decimal) decimal.Decimal {
	if kink.Equal(decimal.Zero) ||
		utilizationRate.LessThanOrEqual(kink) {
		return utilizationRate.Mul(GetMultiplierPerBlock(multiplier)).Add(GetBaseRatePerBlock(baseRate))
	}

	normalRate := kink.Mul(GetMultiplierPerBlock(multiplier)).Add(GetBaseRatePerBlock(baseRate))
	excessUtilRate := utilizationRate.Sub(kink)
	return excessUtilRate.Mul(GetJumpMultiplierPerBlock(jumpMultiplier)).Add(normalRate)
}

// GetSupplyRatePerBlock supply rate per block
func GetSupplyRatePerBlock(utilizationRate, baseRate, multiplier, jumpMultiplier, kink, reserveFactor decimal.Decimal) decimal.Decimal {
	borrowRate := GetBorrowRatePerBlock(utilizationRate, baseRate, multiplier, jumpMultiplier, kink)
	oneMinusReserveFactor := decimal.NewFromInt(1).Sub(reserveFactor)
	rateToPool := borrowRate.Mul(oneMinusReserveFactor)
	return utilizationRate.Mul(rateToPool)
}

// GetBaseRatePerBlock base rate per block
func GetBaseRatePerBlock(baseRate decimal.Decimal) decimal.Decimal {
	return baseRate.Div(BlocksPerYear)
}

// GetMultiplierPerBlock multiplier per block
func GetMultiplierPerBlock(multiplier decimal.Decimal) decimal.Decimal {
	return multiplier.Div(BlocksPerYear)
}

// GetJumpMultiplierPerBlock jump multiplier per block
func GetJumpMultiplierPerBlock(jumpMultiplier decimal.Decimal) decimal.Decimal {
	return jumpMultiplier.Div(BlocksPerYear)
}
