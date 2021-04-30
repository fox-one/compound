package proposal

import (
	"compound/pkg/mtg"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// MarketReq add maret request
type MarketReq struct {
	Symbol               string          `json:"symbol,omitempty"`
	AssetID              string          `json:"asset_id,omitempty"`
	CTokenAssetID        string          `json:"ctoken_asset_id,omitempty"`
	InitExchange         decimal.Decimal `json:"init_exchange,omitempty"`
	ReserveFactor        decimal.Decimal `json:"reserve_factor,omitempty"`
	LiquidationIncentive decimal.Decimal `json:"liquidation_incentive,omitempty"`
	CollateralFactor     decimal.Decimal `json:"collateral_factor,omitempty"`
	BaseRate             decimal.Decimal `json:"base_rate,omitempty"`
	BorrowCap            decimal.Decimal `json:"borrow_cap,omitempty"`
	CloseFactor          decimal.Decimal `json:"close_factor,omitempty"`
	Multiplier           decimal.Decimal `json:"multiplier,omitempty"`
	JumpMultiplier       decimal.Decimal `json:"jump_multiplier,omitempty"`
	Kink                 decimal.Decimal `json:"kink,omitempty"`
}

// MarshalBinary marshal req to binary
func (w MarketReq) MarshalBinary() (data []byte, err error) {
	asset, err := uuid.FromString(w.AssetID)
	if err != nil {
		return nil, err
	}

	ctokenAsset, err := uuid.FromString(w.CTokenAssetID)
	if err != nil {
		return nil, err
	}

	return mtg.Encode(w.Symbol, asset, ctokenAsset, w.InitExchange, w.ReserveFactor, w.LiquidationIncentive, w.CollateralFactor, w.BaseRate, w.BorrowCap, w.CloseFactor, w.Multiplier, w.JumpMultiplier, w.Kink)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *MarketReq) UnmarshalBinary(data []byte) error {
	var symbol string
	var asset, ctokenAsset uuid.UUID
	var initExchange, reserveFactor, liquidationIncentive, collateralFactor, baseRate, borrowCap, closeFactor, multiplier, jumpMultiplier, kink decimal.Decimal

	if _, err := mtg.Scan(data, &symbol, &asset, &ctokenAsset, &initExchange, &reserveFactor, &liquidationIncentive, &collateralFactor, &baseRate, &borrowCap, &closeFactor, &multiplier, &jumpMultiplier, &kink); err != nil {
		return err
	}

	w.Symbol = symbol
	w.AssetID = asset.String()
	w.CTokenAssetID = ctokenAsset.String()
	w.InitExchange = initExchange
	w.ReserveFactor = reserveFactor
	w.LiquidationIncentive = liquidationIncentive
	w.CollateralFactor = collateralFactor
	w.BaseRate = baseRate
	w.BorrowCap = borrowCap
	w.CloseFactor = closeFactor
	w.Multiplier = multiplier
	w.JumpMultiplier = jumpMultiplier
	w.Kink = kink

	return nil
}
