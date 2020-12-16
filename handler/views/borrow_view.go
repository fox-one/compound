package views

import (
	"compound/core"
)

// Borrow supply view
type Borrow struct {
	core.Borrow
	Address string `json:"address"`
}
