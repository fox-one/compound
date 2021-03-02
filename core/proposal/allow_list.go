package proposal

import (
	"compound/pkg/mtg"

	"github.com/gofrs/uuid"
)

// AllowListReq allow list req for add and remove
type AllowListReq struct {
	UserID string `json:"user_id,omitempty"`
	Scope  string `json:"scope,omitempty"`
}

// MarshalBinary marshal allowlist to binary
func (r AllowListReq) MarshalBinary() (data []byte, err error) {
	user, e := uuid.FromString(r.UserID)
	if e != nil {
		return nil, e
	}

	return mtg.Encode(user, r.Scope)
}

// UnmarshalBinary unmarshal allowlist from binary
func (r *AllowListReq) UnmarshalBinary(data []byte) error {
	var user uuid.UUID
	var scope string

	if _, err := mtg.Scan(data, &user, &scope); err != nil {
		return err
	}

	r.UserID = user.String()
	r.Scope = scope

	return nil
}
