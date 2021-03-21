package cmd

import (
	"crypto/ed25519"
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/spf13/cobra"
)

// maintain command
var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "generate ed25519 key pairs",
	Run: func(cmd *cobra.Command, args []string) {
		private := mixin.GenerateEd25519Key()
		public := private.Public().(ed25519.PublicKey)

		cmd.Println("Private", base64.StdEncoding.EncodeToString(private))
		cmd.Println("Public ", base64.StdEncoding.EncodeToString(public))
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}
