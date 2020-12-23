package cmd

import (
	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
