package core

import (
	"compound/internal/compound"
	"context"

	"github.com/shopspring/decimal"
)

// Market market info
type Market struct {
	// asset id
	AssetID string `sql:"size:36;PRIMARY_KEY" json:"asset_id"`
	// 未借出的现金
	TotalCash decimal.Decimal `sql:"type:decimal(24,8)" json:"total_cash"`
	// 已借出的资产
	TotalBorrows decimal.Decimal `sql:"type:decimal(24,8)" json:"total_borrows"`
	// CToken 数量
	CTokens decimal.Decimal `sql:"type:decimal(24,8)" json:"c_tokens"`
	// 平台储备金总量
	TotalReserves decimal.Decimal `sql:"type:decimal(24,8)" json:"total_reserves"`
	// 平台储备金率 (0, 1), 默认为 0.10
	ReserveFactor decimal.Decimal `sql:"type:decimal(24,8)" json:"reserve_factor"`
	// 清算激励因子 (0, 1)
	LiquidationIncentive decimal.Decimal `sql:"type:decimal(24,8)" json:"liquidation_incentive"`
	//抵押因子 = 可借贷价值 / 抵押资产价值，目前compound设置为0.75 
	CollateralFactor decimal.Decimal `sql:"type:decimal(24,8)" json:"collateral_factor"`
	//触发清算因子 [0.05, 0.9]
	CloseFactor decimal.Decimal `sql:"type:decimal(24,8)" json:"close_factor"`
	//基础利率 per year, 0.025
	BaseRate decimal.Decimal `sql:"type:decimal(24,8)" json:"base_rate"`
	// The multiplier of utilization rate that gives the slope of the interest rate. per year
	Multiplier decimal.Decimal `sql:"type:decimal(24,8)" json:"multiplier"`
	// The multiplierPerBlock after hitting a specified utilization point. per year
	JumpMultiplier decimal.Decimal `sql:"type:decimal(24,8)" json:"jump_multiplier"`
	// Kink
	Kink decimal.Decimal `sql:"type:decimal(24,8)" json:"kink"`

}

// BorrowRate get borrow rate
func (m *Market) BorrowRate() decimal.Decimal {
	return compound.GetBorrowRate(m.TotalCash, m.TotalBorrows, m.TotalReserves)
}

// SupplyRate get supply rate
func (m *Market) SupplyRate() decimal.Decimal {
	return compound.GetSupplyRate(m.TotalCash, m.TotalBorrows, m.TotalReserves, m.ReserveFactor)
}

// UtilizationRate utilization rate
func (m *Market) UtilizationRate() decimal.Decimal {
	return compound.UtilizationRate(m.TotalCash, m.TotalBorrows, m.TotalReserves)
}

// ExchangeRate exchange rate
func (m *Market) ExchangeRate() decimal.Decimal {
	return decimal.Zero
}

func (m *Market) Transfer() {

}

func (m *Market) transferAllowed() {

}

// IMarketStore asset store interface
type IMarketStore interface {
	Save(ctx context.Context, market *Market) error
	Find(ctx context.Context, assetID string) (*Market, error)
	All(ctx context.Context) ([]*Market, error)
}

// IMarket market interface
type IMarket interface {
	BorrowRatePerBlock() decimal.Decimal
	SupplyRatePerBlock() decimal.Decimal
	UtilizationRate() decimal.Decimal
	ExchangeRate() decimal.Decimal
	Mint() 
	MintAllowed()
}
