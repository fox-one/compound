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
	CollateralFactor     decimal.Decimal `json:"collateral_factor,omitempty"`
	BaseRate             decimal.Decimal `json:"base_rate,omitempty"`
}

// MarshalBinary marshal req to binary
func (w UpdateMarketReq) MarshalBinary() (data []byte, err error) {
	return mtg.Encode(w.Symbol, w.InitExchange, w.ReserveFactor, w.LiquidationIncentive, w.CollateralFactor, w.BaseRate)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *UpdateMarketReq) UnmarshalBinary(data []byte) error {
	var symbol string
	var initExchange, reserveFactor, liquidationIncentive, collateralFactor, baseRate decimal.Decimal

	if _, err := mtg.Scan(data, &symbol, &initExchange, &reserveFactor, &liquidationIncentive, &collateralFactor, &baseRate); err != nil {
		return err
	}

	w.Symbol = symbol
	w.InitExchange = initExchange
	w.ReserveFactor = reserveFactor
	w.LiquidationIncentive = liquidationIncentive
	w.CollateralFactor = collateralFactor
	w.BaseRate = baseRate

	return nil
}

// UpdateMarketAdvanceReq req of market advance parameters
type UpdateMarketAdvanceReq struct {
	Symbol         string          `json:"symbol,omitempty"`
	BorrowCap      decimal.Decimal `json:"borrow_cap,omitempty"`
	CloseFactor    decimal.Decimal `json:"close_factor,omitempty"`
	Multiplier     decimal.Decimal `json:"multiplier,omitempty"`
	JumpMultiplier decimal.Decimal `json:"jump_multiplier,omitempty"`
	Kink           decimal.Decimal `json:"kink,omitempty"`
}

// MarshalBinary marshal req to binary
func (w UpdateMarketAdvanceReq) MarshalBinary() (data []byte, err error) {
	return mtg.Encode(w.Symbol, w.BorrowCap, w.CloseFactor, w.Multiplier, w.JumpMultiplier, w.Kink)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *UpdateMarketAdvanceReq) UnmarshalBinary(data []byte) error {
	var symbol string
	var borrowCap, closeFactor, multiplier, jumpMultiplier, kink decimal.Decimal

	if _, err := mtg.Scan(data, &symbol, &borrowCap, &closeFactor, &multiplier, &jumpMultiplier, &kink); err != nil {
		return err
	}

	w.Symbol = symbol
	w.BorrowCap = borrowCap
	w.CloseFactor = closeFactor
	w.Multiplier = multiplier
	w.JumpMultiplier = jumpMultiplier
	w.Kink = kink

	return nil
}
