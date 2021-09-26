package core

import (
	"crypto/ed25519"

	"github.com/asaskevich/govalidator"
	"github.com/shopspring/decimal"
)

const (
	SysVersion int64 = 1
)

type (
	// System stores system information.
	System struct {
		Admins     []string
		ClientID   string
		Members    []*Member
		MemberIDs  []string
		Threshold  uint8
		VoteAsset  string
		VoteAmount decimal.Decimal
		PrivateKey ed25519.PrivateKey
		SignKey    ed25519.PrivateKey
		Genesis    int64
		Version    string
	}
)

func (s *System) IsMember(id string) bool {
	return govalidator.IsIn(id, s.MemberIDs...)
}

func (s *System) IsStaff(id string) bool {
	return govalidator.IsIn(id, s.Admins...)
}
