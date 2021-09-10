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
	PriceThreshold       int             `json:"price_threshold,omitempty"`
	Price                decimal.Decimal `json:"price,omitempty"`
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

	return mtg.Encode(
		w.Symbol,
		asset,
		ctokenAsset,
		w.InitExchange,
		w.ReserveFactor,
		w.LiquidationIncentive,
		w.CollateralFactor,
		w.BaseRate,
		w.BorrowCap,
		w.CloseFactor,
		w.Multiplier,
		w.JumpMultiplier,
		w.Kink,
		w.PriceThreshold,
		w.Price,
	)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *MarketReq) UnmarshalBinary(data []byte) error {
	var (
		req           MarketReq
		assetID       uuid.UUID
		ctokenAssetID uuid.UUID
	)
	if _, err := mtg.Scan(data,
		&req.Symbol,
		&assetID,
		&ctokenAssetID,
		&req.InitExchange,
		&req.ReserveFactor,
		&req.LiquidationIncentive,
		&req.CollateralFactor,
		&req.BaseRate,
		&req.BorrowCap,
		&req.CloseFactor,
		&req.Multiplier,
		&req.JumpMultiplier,
		&req.Kink,
		&req.PriceThreshold,
		&req.Price,
	); err != nil {
		return err
	}
	req.AssetID = assetID.String()
	req.CTokenAssetID = ctokenAssetID.String()

	*w = req
	return nil
}
