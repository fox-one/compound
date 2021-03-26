package main

import (
	"compound/cmd"
	"compound/core"
	"fmt"
	"time"
)

var (
	version string
	commit  string
)

func main() {
	version := fmt.Sprintf("%s-%s", version, commit)

	fmt.Println(core.CalculatePriceTick(time.Now()))

	cmd.Run(version)
}
