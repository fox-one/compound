package cmd

import (
	"compound/core"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func update(m *core.Market) {
	m.Price = decimal.NewFromInt(1200)
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
