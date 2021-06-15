package cmd

import (
	"crypto/ed25519"
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/pandodao/blst"
	"github.com/spf13/cobra"
)

// maintain command
var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "generate ed25519 or blst key pairs by the flag 'cipher'",
	Run: func(cmd *cobra.Command, args []string) {
		cipher, err := cmd.Flags().GetString("cipher")
		if err != nil {
			panic(err)
		}

		cmd.Println("cipher is:", cipher)

		if cipher == "blst" {
			private := blst.GenerateKey()
			public := private.PublicKey()

			cmd.Println("blst private key: ", private.String())
			cmd.Println("blst public key:", public.String())
		} else {
			private := mixin.GenerateEd25519Key()
			public := private.Public().(ed25519.PublicKey)

			cmd.Println("ed25519 Private key: ", base64.StdEncoding.EncodeToString(private))
			cmd.Println("ed25519 Public key: ", base64.StdEncoding.EncodeToString(public))
		}

	},
}

func init() {
	rootCmd.AddCommand(keysCmd)

	keysCmd.Flags().String("cipher", "ed25519", "cipher type")
}
