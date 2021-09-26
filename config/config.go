package config

import (
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

// Config compound config
type (
	Config struct {
		Genesis int64     `json:"genesis"`
		DB      db.Config `json:"db"`
		Dapp    Dapp      `json:"dapp"`
		Group   Group     `json:"group"`
		DataDog DataDog   `json:"data_dog"`
	}

	// Group group config
	Group struct {
		PrivateKey string       `json:"private_key"`
		SignKey    string       `json:"sign_key"`
		Admins     []string     `json:"admins"`
		Threshold  uint8        `json:"threshold"`
		Members    []MemberConf `json:"members"`
		Vote       Vote         `json:"vote"`
	}

	// MemberConf member info
	MemberConf struct {
		ClientID  string `json:"client_id"`
		VerifyKey string `json:"verify_key"`
	}

	// Vote vote config info
	Vote struct {
		Asset  string          `json:"asset"`
		Amount decimal.Decimal `json:"amount"`
	}

	// Dapp mixin dapp config
	Dapp struct {
		mixin.Keystore
		Pin string `json:"pin"`
	}

	DataDog struct {
		ConversationID string `json:"conversation_id,omitempty"`
	}
)

func defaultVote(cfg *Config) {
	if cfg.Group.Vote.Asset == "" {
		cfg.Group.Vote.Asset = "965e5c6e-434c-3fa9-b780-c50f43cd955c"
	}

	if cfg.Group.Vote.Amount.IsZero() {
		cfg.Group.Vote.Amount = decimal.New(1, -8)
	}
}
