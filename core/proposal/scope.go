package proposal

import (
	"compound/pkg/mtg"
)

// ScopeReq allow list scope req for add and remove
type ScopeReq struct {
	Scope string `json:"scope,omitempty"`
}

// MarshalBinary marshal allowlist scope to binary
func (r ScopeReq) MarshalBinary() (data []byte, err error) {
	return mtg.Encode(r.Scope)
}

// UnmarshalBinary unmarshal allowlist scope from binary
func (r *ScopeReq) UnmarshalBinary(data []byte) error {
	var scope string

	if _, err := mtg.Scan(data, &scope); err != nil {
		return err
	}

	r.Scope = scope

	return nil
}
