package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

const (
	_ MarketStatus = iota
	// MarketStatusOpen open
	MarketStatusOpen
	// MarketStatusClose close
	MarketStatusClose
)

type (
	// MarketStatus market status
	MarketStatus int

	// Market market info
	Market struct {
		ID            uint64          `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
		AssetID       string          `sql:"size:36;unique_index:asset_idx" json:"asset_id"`
		Symbol        string          `sql:"size:20;unique_index:symbol_idx" json:"symbol"`
		CTokenAssetID string          `sql:"size:36;unique_index:ctoken_asset_idx" json:"ctoken_asset_id"`
		TotalCash     decimal.Decimal `sql:"type:decimal(32,16)" json:"total_cash"`
		TotalBorrows  decimal.Decimal `sql:"type:decimal(32,16)" json:"total_borrows"`
		MaxPledge     decimal.Decimal `sql:"type:decimal(32,16)" json:"max_pledge"`
		// 保留金
		Reserves decimal.Decimal `sql:"type:decimal(32,16)" json:"reserves"`
		// CToken 累计铸造出来的币的数量
		CTokens decimal.Decimal `sql:"type:decimal(32,16)" json:"ctokens"`
		// 初始兑换率
		InitExchangeRate decimal.Decimal `sql:"type:decimal(32,16);default:0" json:"init_exchange_rate"`
		// 平台保留金率 (0, 1), 默认为 0.10
		ReserveFactor decimal.Decimal `sql:"type:decimal(32,16)" json:"reserve_factor"`
		// 清算激励因子 (0, 1), 一般为0.1
		LiquidationIncentive decimal.Decimal `sql:"type:decimal(32,16)" json:"liquidation_incentive"`
		// 资金池的最小资金量
		BorrowCap decimal.Decimal `sql:"type:decimal(32,16);default:0" json:"borrow_cap"`
		//抵押因子 = 可借贷价值 / 抵押资产价值，目前compound设置为0.75. 稳定币(USDT)的抵押率是0,即不可抵押
		CollateralFactor decimal.Decimal `sql:"type:decimal(32,16)" json:"collateral_factor"`
		//触发清算因子 [0.05, 0.9] 清算人最大可清算的资产比例
		CloseFactor decimal.Decimal `sql:"type:decimal(32,16)" json:"close_factor"`
		//基础利率 per year, 0.025
		BaseRate decimal.Decimal `sql:"type:decimal(32,16)" json:"base_rate"`
		// The multiplier of utilization rate that gives the slope of the interest rate. per year
		Multiplier decimal.Decimal `sql:"type:decimal(32,16)" json:"multiplier"`
		// The multiplierPerBlock after hitting a specified utilization point. per year
		JumpMultiplier decimal.Decimal `sql:"type:decimal(32,16)" json:"jump_multiplier"`
		// Kink
		Kink decimal.Decimal `sql:"type:decimal(32,16)" json:"kink"`
		//当前区块高度
		BlockNumber        int64           `json:"block_number"`
		UtilizationRate    decimal.Decimal `sql:"type:decimal(32,16)" json:"utilization_rate"`
		ExchangeRate       decimal.Decimal `sql:"type:decimal(32,16)" json:"exchange_rate"`
		SupplyRatePerBlock decimal.Decimal `sql:"type:decimal(32,16)" json:"supply_rate_per_block"`
		BorrowRatePerBlock decimal.Decimal `sql:"type:decimal(32,16)" json:"borrow_rate_per_block"`
		Price              decimal.Decimal `sql:"type:decimal(32,16)" json:"price"`
		PriceThreshold     int             `json:"price_threshold"`
		PriceUpdatedAt     time.Time       `json:"price_updated_at"`
		BorrowIndex        decimal.Decimal `sql:"type:decimal(28,16)" json:"borrow_index"`
		Version            int64           `sql:"default:0" json:"version"`
		Status             MarketStatus    `sql:"default:1" json:"status"`
		CreatedAt          time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"created_at"`
		UpdatedAt          time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	}

	// IMarketStore asset store interface
	IMarketStore interface {
		Create(ctx context.Context, market *Market) error
		Find(ctx context.Context, assetID string) (*Market, error)
		FindBySymbol(ctx context.Context, symbol string) (*Market, error)
		FindByCToken(ctx context.Context, ctokenAssetID string) (*Market, error)
		All(ctx context.Context) ([]*Market, error)
		AllAsMap(ctx context.Context) (map[string]*Market, error)
		Update(ctx context.Context, market *Market, version int64) error
	}
)

// IsValid is valid status
func (s MarketStatus) IsValid() bool {
	return s == MarketStatusClose ||
		s == MarketStatusOpen
}

func (m Market) Format() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return bytes
}

func (m Market) IsMarketClosed() bool {
	return m.Status == MarketStatusClose
}

// BorrowAllowed check borrow capacity, check account liquidity
func (m Market) BorrowAllowed(borrowAmount decimal.Decimal) bool {
	if !borrowAmount.IsPositive() {
		return false
	}

	// check borrow cap
	balance := m.TotalCash.Sub(m.Reserves)
	if balance.LessThan(m.BorrowCap) {
		return false
	}

	if borrowAmount.GreaterThan(balance.Sub(m.BorrowCap)) {
		return false
	}

	return true
}

func (m Market) RedeemAllowed(redeemTokens decimal.Decimal) bool {
	amount := redeemTokens.Mul(m.ExchangeRate)
	supplies := m.TotalCash.Sub(m.Reserves)
	return supplies.GreaterThan(amount)
}

func (m Market) CurExchangeRate() decimal.Decimal {
	if v := m.ExchangeRate; v.IsPositive() {
		return v
	} else {
		return m.InitExchangeRate
	}
}
