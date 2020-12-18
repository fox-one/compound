package core

import (
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// Config compound config
type Config struct {
	Genesis     int64       `json:"genesis"`
	Location    string      `json:"location"`
	DB          db.Config   `json:"db"`
	Dapp        Dapp        `json:"dapp"`
	Group       Group       `json:"group"`
	PriceOracle PriceOracle `json:"price_oracle"`
}

// IsAdmin check if the user is admin
func (c *Config) IsAdmin(userID string) bool {
	if len(c.Group.Admins) <= 0 {
		return false
	}

	for _, a := range c.Group.Admins {
		if a == userID {
			return true
		}
	}

	return false
}

// Group group config
type Group struct {
	PrivateKey string       `json:"private_key"`
	SignKey    string       `json:"sign_key"`
	Admins     []string     `json:"admins"`
	Threshold  uint8        `json:"threshold"`
	Members    []MemberConf `json:"members"`
	Vote       Vote         `json:"vote"`
}

// MemberConf member info
type MemberConf struct {
	ClientID  string `json:"client_id"`
	VerifyKey string `json:"verify_key"`
}

// Vote vote config info
type Vote struct {
	Asset  string          `json:"asset"`
	Amount decimal.Decimal `json:"amount"`
}

// Dapp mixin dapp config
type Dapp struct {
	mixin.Keystore
	ClientSecret string `json:"client_secret"`
	Pin          string `json:"pin"`
}

// PriceOracle price oracle config
type PriceOracle struct {
	EndPoint string `json:"end_point"`
}
