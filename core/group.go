package core

import (
	"compound/pkg/mtg"
	"crypto/ed25519"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
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
func DecodeMemberProposalTransactionAction(message []byte, members []*Member) (*Member, ActionType, []byte, error) {
	member, action, data, e := decodeSignedPackedProposalData(message, members)
	if e != nil {
		return decodeRawProposalData(message, members)
	}

	return member, action, data, nil
}

func decodeSignedPackedProposalData(message []byte, members []*Member) (*Member, ActionType, []byte, error) {
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

func decodeRawProposalData(message []byte, members []*Member) (*Member, ActionType, []byte, error) {
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

	return nil, action, content, nil
}
