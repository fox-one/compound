package core

import (
	"github.com/pandodao/blst"
)

type (
	Signer struct {
		Index     uint64          `json:"index,omitempty"`
		VerifyKey *blst.PublicKey `json:"verify_key,omitempty"`
	}

	Receiver struct {
		Members   []string `json:"members,omitempty"`
		Threshold uint8    `json:"threshold"`
	}

	Asset struct {
		AssetID string `json:"asset_id,omitempty"`
		Symbol  string `json:"symbol,omitempty"`
	}

	PriceRequest struct {
		Asset

		TraceID   string    `json:"trace_id,omitempty"`
		Receiver  *Receiver `json:"receiver,omitempty"`
		Signers   []*Signer `json:"signers,omitempty"`
		Threshold uint8     `json:"threshold"`
	}
)
