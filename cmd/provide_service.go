package cmd

import (
	"compound/core"

	"github.com/fox-one/mixin-sdk-go"
)

func provideMixinClient() *mixin.Client {
	c, err := mixin.NewFromKeystore(&cfg.Mixin.Keystore)
	if err != nil {
		panic(err)
	}

	return c
}

func provideConfig() *core.Config {
	return &cfg
}
