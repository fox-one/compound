package core

import (
	"compound/pkg/mtg"
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
)

// Member member
type Member struct {
	ClientID  string
	Name      string
	VerifyKey ed25519.PublicKey
}

func DecodeTransactionAction(privateKey ed25519.PrivateKey, message []byte) (ActionType, []byte, error) {
	b, err := mtg.Decrypt(message, privateKey)
	if err != nil {
		return 0, nil, err
	}

	var t int
	b, err = mtg.Scan(b, &t)
	if err != nil {
		return 0, nil, err
	}

	typ := ParseActionType(ActionType(t).String())
	if typ == 0 {
		return 0, nil, fmt.Errorf("invalid transaction type %d", t)
	}

	return typ, b, nil
}

// DecodeMemberActionV1 decode member vote transaction
func DecodeMemberActionV1(message []byte, members []*Member) (*Member, []byte, error) {
	body, sig, err := mtg.Unpack(message)
	if err != nil {
		return nil, nil, err
	}

	var id uuid.UUID
	content, err := mtg.Scan(body, &id)
	if err != nil {
		return nil, nil, err
	}

	for _, member := range members {
		if member.ClientID != id.String() {
			continue
		}

		if !mtg.Verify(body, sig, member.VerifyKey) {
			return nil, nil, errors.New("verify sig failed")
		}

		return member, content, nil
	}

	return nil, nil, errors.New("member not found")
}

// Deprecated since sysver 1
//	high risks, did not verify the node's signature
// DecodeMemberActionV0 decode member vote transaction
func DecodeMemberActionV0(message []byte, members []*Member) (*Member, ActionType, []byte, error) {
	member, action, data, e := decodeMemberActionV0(message, members)
	if e != nil {
		return decodeMemberActionUnsecureV0(message, members)
	}

	return member, action, data, nil
}

func decodeMemberActionV0(message []byte, members []*Member) (*Member, ActionType, []byte, error) {
	body, sig, err := mtg.Unpack(message)
	if err != nil {
		return nil, ActionTypeDefault, nil, err
	}

	var clientID uuid.UUID
	var actionType int
	content, err := mtg.Scan(body, &clientID, &actionType)
	if err != nil {
		return nil, ActionTypeDefault, nil, err
	}

	action := ActionType(actionType)
	if !action.IsProposalAction() {
		return nil, ActionTypeDefault, nil, errors.New("invalid proposal action")
	}

	for _, m := range members {
		if m.ClientID != clientID.String() {
			continue
		}

		if !mtg.Verify(body, sig, m.VerifyKey) {
			return nil, action, nil, errors.New("verify sig failed")
		}
		return m, action, content, nil
	}

	return nil, action, nil, errors.New("member not found")
}

func decodeMemberActionUnsecureV0(message []byte, members []*Member) (*Member, ActionType, []byte, error) {
	var clientID uuid.UUID
	var actionType int
	content, err := mtg.Scan(message, &clientID, &actionType)
	if err != nil {
		return nil, ActionTypeDefault, nil, err
	}

	action := ActionType(actionType)
	if !action.IsProposalAction() {
		return nil, ActionTypeDefault, nil, errors.New("invalid proposal action")
	}

	for _, m := range members {
		if m.ClientID != clientID.String() {
			continue
		}

		return m, action, content, nil
	}

	return nil, action, content, errors.New("member not found")
}
