package proposal

import (
	"compound/pkg/mtg"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// ProvidePriceReq provide price request
type ProvidePriceReq struct {
	AssetID string          `json:"asset_id,omitempty"`
	Price   decimal.Decimal `json:"price,omitempty"`
}

// MarshalBinary marshal price to binary
func (p ProvidePriceReq) MarshalBinary() (data []byte, err error) {
	asset, e := uuid.FromString(p.AssetID)
	if e != nil {
		return nil, e
	}
	return mtg.Encode(asset, p.Price)
}

// UnmarshalBinary unmarshal bytes to price
func (p *ProvidePriceReq) UnmarshalBinary(data []byte) error {
	var asset uuid.UUID
	var price decimal.Decimal
	if _, err := mtg.Scan(data, &asset, &price); err != nil {
		return err
	}

	p.AssetID = asset.String()
	p.Price = price

	return nil
}
