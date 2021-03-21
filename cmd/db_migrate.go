package cmd

import (
	"github.com/fox-one/pkg/store/db"
	"github.com/spf13/cobra"
)

// command for migrating database
var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"setdb"},
	Short:   "migrate database tables",
	Run: func(cmd *cobra.Command, args []string) {
		database := provideDatabase()
		defer database.Close()

		if err := db.Migrate(database); err != nil {
			cmd.PrintErrln("migrate database error:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
