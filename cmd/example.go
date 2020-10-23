package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(time.Now().UTC().Unix())
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
