package proposal

import (
	"compound/core"
	"compound/pkg/mtg"
	"errors"

	"github.com/gofrs/uuid"
)

// MarketStatusReq update maret status as open or close
type MarketStatusReq struct {
	AssetID string            `json:"asset_id,omitempty"`
	Status  core.MarketStatus `json:"status,omitempty"`
}

// MarshalBinary marshal req to binary
func (w MarketStatusReq) MarshalBinary() (data []byte, err error) {
	asset, err := uuid.FromString(w.AssetID)
	if err != nil {
		return nil, err
	}

	return mtg.Encode(asset, w.Status)
}

// UnmarshalBinary unmarshal bytes to withdraw
func (w *MarketStatusReq) UnmarshalBinary(data []byte) error {
	var asset uuid.UUID
	var status int

	if _, err := mtg.Scan(data, &asset, &status); err != nil {
		return err
	}

	s := core.MarketStatus(status)
	if !s.IsValid() {
		return errors.New("invalid market status")
	}

	w.AssetID = asset.String()
	w.Status = s

	return nil
}
