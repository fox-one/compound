package cmd

import (
	"compound/core"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		var transferAction core.TransferAction
		m, e := base64.StdEncoding.DecodeString("eyJzIjo1LCJmIjoiYjI4ODVkNDctMzkzZS00MmM2LTg2YWQtYjI2MDgzZDJiYmYzIn0=")
		e = json.Unmarshal(m, &transferAction)
		if e != nil {
			panic(e)
		}

		fmt.Println(transferAction)
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
