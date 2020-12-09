package core

import (
	"compound/pkg/mtg"
	"crypto/ed25519"
	"errors"

	"github.com/gofrs/uuid"
)

// Member member
type Member struct {
	ClientID  string
	Name      string
	VerifyKey ed25519.PublicKey
}

// DecodeUserTransactionAction decode user transaction
func DecodeUserTransactionAction(privateKey ed25519.PrivateKey, message []byte) (ActionType, []byte, error) {
	b, err := mtg.Decrypt(message, privateKey)
	if err != nil {
		return 0, nil, err
	}

	var t int
	b, err = mtg.Scan(b, &t)
	if err != nil {
		return 0, nil, err
	}

	return ActionType(t), b, nil
}

// DecodeMemberProposalTransactionAction decode member vote transaction
func DecodeMemberProposalTransactionAction(message []byte, members []*Member) (*Member, []byte, error) {
	body, sig, err := mtg.Unpack(message)
	if err != nil {
		return nil, nil, err
	}

	var id uuid.UUID
	content, err := mtg.Scan(body, &id)
	if err != nil {
		return nil, nil, err
	}

	for _, m := range members {
		if m.ClientID != id.String() {
			continue
		}

		if !mtg.Verify(body, sig, m.VerifyKey) {
			return nil, nil, errors.New("verify sig failed")
		}
		return m, content, nil
	}

	return nil, nil, errors.New("member not found")
}
