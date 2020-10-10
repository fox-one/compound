package core

import (
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/store/db"
)

// Config compound config
type Config struct {
	DB    db.Config `json:"db"`
	Redis Redis     `json:"redis,omitempty"`
	Mixin Mixin     `json:"mixin"`
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
