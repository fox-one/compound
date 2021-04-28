package core

import (
	"crypto/ed25519"

	"github.com/shopspring/decimal"
)

// System stores system information.
type System struct {
	Admins         []string
	ClientID       string
	ClientSecret   string
	Members        []*Member
	Threshold      uint8
	VoteAsset      string
	VoteAmount     decimal.Decimal
	PrivateKey     ed25519.PrivateKey
	SignKey        ed25519.PrivateKey
	PriceThreshold uint8
	Location       string
	Genesis        int64
	Version        string
}

// MemberIDs member ids
func (s *System) MemberIDs() []string {
	ids := make([]string, len(s.Members))
	for idx, m := range s.Members {
		ids[idx] = m.ClientID
	}

	return ids
}

// IsAdmin is admin
func (s *System) IsAdmin(userID string) bool {
	if len(s.Admins) == 0 {
		return false
	}

	for _, a := range s.Admins {
		if a == userID {
			return true
		}
	}

	return false
}
