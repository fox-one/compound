package proposal

import (
	"compound/pkg/mtg"

	"github.com/gofrs/uuid"
)

// AddMarketReq add maret request
type AddMarketReq struct {
	Symbol        string `json:"symbol,omitempty"`
	AssetID       string `json:"asset_id,omitempty"`
	CTokenAssetID string `json:"ctoken_asset_id,omitempty"`
}

// MarshalBinary marshal req to binary
func (w AddMarketReq) MarshalBinary() (data []byte, err error) {
	symbol := w.Symbol

	asset, err := uuid.FromString(w.AssetID)
	if err != nil {
		return nil, err
	}

	ctokenAsset, err := uuid.FromString(w.CTokenAssetID)
	if err != nil {
		return nil, err
	}

	return mtg.Encode(symbol, asset, ctokenAsset)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *AddMarketReq) UnmarshalBinary(data []byte) error {
	var symbol string
	var asset, ctokenAsset uuid.UUID

	if _, err := mtg.Scan(data, &symbol, &asset, &ctokenAsset); err != nil {
		return err
	}

	w.Symbol = symbol
	w.AssetID = asset.String()
	w.CTokenAssetID = ctokenAsset.String()

	return nil
}
