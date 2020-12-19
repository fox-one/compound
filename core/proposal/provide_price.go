package proposal

import (
	"compound/pkg/mtg"

	"github.com/shopspring/decimal"
)

// ProvidePriceReq provide price request
type ProvidePriceReq struct {
	Symbol string          `json:"symbol,omitempty"`
	Price  decimal.Decimal `json:"price,omitempty"`
}

// MarshalBinary marshal price to binary
func (p ProvidePriceReq) MarshalBinary() (data []byte, err error) {
	return mtg.Encode(p.Symbol, p.Price)
}

// UnmarshalBinary unmarshal bytes to price
func (p *ProvidePriceReq) UnmarshalBinary(data []byte) error {
	var symbol string
	var price decimal.Decimal
	if _, err := mtg.Scan(data, &symbol, &price); err != nil {
		return err
	}

	p.Symbol = symbol
	p.Price = price

	return nil
}
