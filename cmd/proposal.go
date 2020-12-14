package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

var proposalCmd = &cobra.Command{
	Use:     "proposal<command>",
	Aliases: []string{"pp"},
	Short:   "Manager proposals",
	Example: heredoc.Doc(``),
}

func init() {
	rootCmd.AddCommand(proposalCmd)
}
