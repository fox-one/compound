package views

import (
	"compound/core"
)

// Supply supply view
type Supply struct {
	core.Supply
	Address string `json:"address"`
}
