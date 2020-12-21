package proposal

import (
	"compound/pkg/mtg"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// WithdrawReq withdraw request info
type WithdrawReq struct {
	Opponent string          `json:"opponent,omitempty"`
	Asset    string          `json:"asset,omitempty"`
	Amount   decimal.Decimal `json:"amount,omitempty"`
}

// MarshalBinary marshal Withdraw to binary
func (w WithdrawReq) MarshalBinary() (data []byte, err error) {
	opponent, err := uuid.FromString(w.Opponent)
	if err != nil {
		return nil, err
	}

	asset, err := uuid.FromString(w.Asset)
	if err != nil {
		return nil, err
	}

	return mtg.Encode(opponent, asset, w.Amount)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *WithdrawReq) UnmarshalBinary(data []byte) error {
	var opponent, asset uuid.UUID
	var amount decimal.Decimal
	if _, err := mtg.Scan(data, &opponent, &asset, &amount); err != nil {
		return err
	}

	w.Opponent = opponent.String()
	w.Asset = asset.String()
	w.Amount = amount

	return nil
}
