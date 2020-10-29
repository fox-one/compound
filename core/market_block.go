package core

import "github.com/shopspring/decimal"

// MarketBlock market info per block
type MarketBlock struct {
	ID              uint64          `sql:"AUTO_INCREMENT;PRIMARY_KEY" json:"id"`
	Symbol          string          `sql:"size:20" json:"symbol"`
	AssetID         string          `sql:"size:36;unique_index:asset_block_idx" json:"asset_id"`
	Block           int64           `sql:"unique_index:asset_block_idx" json:"block"`
	UtilizationRate decimal.Decimal `sql:"type:decimal(20,8)" json:"utilization_rate"`
	ExchangeRate    decimal.Decimal `sql:"type:decimal(20,8)" json:"exchange_rate"`
	BorrowRate      decimal.Decimal `sql:"type:decimal(20,8)" json:"borrow_rate"`
	SupplyRate      decimal.Decimal `sql:"type:decimal(20,8)" json:"supply_rate"`
}
