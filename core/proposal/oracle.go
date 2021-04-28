package proposal

import (
	"compound/pkg/mtg"
	"encoding/base64"

	"github.com/gofrs/uuid"
)

type AddOracleSignerReq struct {
	UserID    string `json:"user_id,omitempty"`
	PublicKey string `json:"publick_key,omitempty"`
}

type RemoveOracleSignerReq struct {
	UserID string `json:"user_id,omitempty"`
}

// MarshalBinary marshal req to binary
func (r AddOracleSignerReq) MarshalBinary() (data []byte, err error) {
	user, err := uuid.FromString(r.UserID)
	if err != nil {
		return nil, err
	}

	pk, err := base64.StdEncoding.DecodeString(r.PublicKey)
	if err != nil {
		return nil, err
	}

	return mtg.Encode(user, mtg.RawMessage(pk))
}

// UnmarshalBinary unmarshal bytes
func (r *AddOracleSignerReq) UnmarshalBinary(data []byte) error {
	var user uuid.UUID
	var publicKey mtg.RawMessage

	if _, err := mtg.Scan(data, &user, &publicKey); err != nil {
		return err
	}

	r.UserID = user.String()
	r.PublicKey = base64.StdEncoding.EncodeToString(publicKey)

	return nil
}

// MarshalBinary marshal req to binary
func (r RemoveOracleSignerReq) MarshalBinary() (data []byte, err error) {
	user, err := uuid.FromString(r.UserID)
	if err != nil {
		return nil, err
	}

	return mtg.Encode(user)
}

// UnmarshalBinary unmarshal bytes
func (r *RemoveOracleSignerReq) UnmarshalBinary(data []byte) error {
	var user uuid.UUID

	if _, err := mtg.Scan(data, &user); err != nil {
		return err
	}

	r.UserID = user.String()

	return nil
}
