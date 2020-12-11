package proposal

import (
	"compound/pkg/mtg"

	"github.com/shopspring/decimal"
)

// UpdateMarketReq update market request
type UpdateMarketReq struct {
	Symbol               string          `json:"symbol,omitempty"`
	InitExchange         decimal.Decimal `json:"init_exchange,omitempty"`
	ReserveFactor        decimal.Decimal `json:"reserve_factor,omitempty"`
	LiquidationIncentive decimal.Decimal `json:"liquidation_incentive,omitempty"`
	BorrowCap            decimal.Decimal `json:"borrow_cap,omitempty"`
	CollateralFactor     decimal.Decimal `json:"collateral_factor,omitempty"`
	CloseFactor          decimal.Decimal `json:"close_factor,omitempty"`
	BaseRate             decimal.Decimal `json:"base_rate,omitempty"`
	Multiplier           decimal.Decimal `json:"multiplier,omitempty"`
	JumpMultiplier       decimal.Decimal `json:"jump_multiplier,omitempty"`
	Kink                 decimal.Decimal `json:"kink,omitempty"`
}

// MarshalBinary marshal req to binary
func (w UpdateMarketReq) MarshalBinary() (data []byte, err error) {
	return mtg.Encode(w.Symbol, w.InitExchange, w.ReserveFactor, w.LiquidationIncentive, w.BorrowCap, w.CollateralFactor, w.CloseFactor, w.BaseRate, w.Multiplier, w.JumpMultiplier, w.Kink)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *UpdateMarketReq) UnmarshalBinary(data []byte) error {
	var symbol string
	var initExchange, reserveFactor, liquidationIncentive, borrowCap, collateralFactor, closeFactor, baseRate, multiplier, jumpMultiplier, kink decimal.Decimal

	if _, err := mtg.Scan(data, &symbol, &initExchange, &reserveFactor, &liquidationIncentive, &borrowCap, &collateralFactor, &closeFactor, &baseRate, &multiplier, &jumpMultiplier, &kink); err != nil {
		return err
	}

	w.Symbol = symbol
	w.InitExchange = initExchange
	w.ReserveFactor = reserveFactor
	w.LiquidationIncentive = liquidationIncentive
	w.BorrowCap = borrowCap
	w.CollateralFactor = collateralFactor
	w.CloseFactor = closeFactor
	w.BaseRate = baseRate
	w.Multiplier = multiplier
	w.JumpMultiplier = jumpMultiplier
	w.Kink = kink

	return nil
}
