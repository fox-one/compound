package core

import (
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
)

// Config compound config
type Config struct {
	App         App         `json:"app"`
	DB          db.Config   `json:"db"`
	MainWallet  MainWallet  `json:"main_wallet"`
	GasWallet   GasWallet   `json:"gas_wallet"`
	PriceOracle PriceOracle `json:"price_oracle"`
	Admins      []string    `json:"admins"`
}

// IsAdmin check if the user is admin
func (c *Config) IsAdmin(userID string) bool {
	if len(c.Admins) <= 0 {
		return false
	}

	for _, a := range c.Admins {
		if a == userID {
			return true
		}
	}

	return false
}

// MainWallet mixin dapp config
type MainWallet struct {
	mixin.Keystore
	ClientSecret string `json:"client_secret"`
	Pin          string `json:"pin"`
}

// App app config
type App struct {
	AESKey     string `json:"aes_key"`
	GasAssetID string `json:"gas_asset_id"`
	Genesis    int64  `json:"genesis"`
	Location   string `json:"location"`
}

// GasWallet gas wallet
type GasWallet struct {
	mixin.Keystore
	ClientSecret string `json:"client_secret"`
	Pin          string `json:"pin"`
}

// PriceOracle price oracle config
type PriceOracle struct {
	EndPoint string `json:"end_point"`
}
