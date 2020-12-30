package cmd

import (
	"compound/core"
	"fmt"

	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		modifier := fmt.Sprintf("%s.%d", "e15a8248-cd64-4e28-b10d-4907236e9fca", core.ActionTypeBorrow)

		fmt.Println(modifier)
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
