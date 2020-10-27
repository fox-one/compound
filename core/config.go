package core

import (
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
)

// Config compound config
type Config struct {
	DB          db.Config   `json:"db"`
	Redis       Redis       `json:"redis,omitempty"`
	Mixin       Mixin       `json:"mixin"`
	App         App         `json:"app"`
	BlockWallet BlockWallet `json:"block_wallet"`
}

// Redis redis config
type Redis struct {
	Addr string `json:"addr,omitempty"`
	DB   int    `json:"db,omitempty"`
}

// Mixin mixin dapp config
type Mixin struct {
	mixin.Keystore
	ClientSecret string `json:"client_secret"`
	Pin          string `json:"pin"`
}

// App app config
type App struct {
	AESKey          string `json:"aes_key"`
	BlockAssetID    string `json:"block_asset_id"`
	SecondsPerBlock int64  `json:"seconds_per_block"`
	Genesis         int64  `json:"genesis"`
	Location        string `json:"location"`
}

// BlockWallet block wallet
type BlockWallet struct {
	mixin.Keystore
	ClientSecret string `json:"client_secret"`
	Pin          string `json:"pin"`
}
