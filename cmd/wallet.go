package cmd

import (
	"encoding/json"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/spf13/cobra"
)

var walletCmd = &cobra.Command{
	Use:     "new-wallet",
	Aliases: []string{"nw"},
	Short:   "new wallet",
	Run: func(cmd *cobra.Command, args []string) {
		walletService := provideWalletService()
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			cmd.PrintErrln(err)
			return
		}

		pin := mixin.RandomPin()
		keystore, _, err := walletService.NewWallet(cmd.Context(), name, pin)
		if err != nil {
			cmd.PrintErrln(err)
			return
		}

		var data struct {
			*mixin.Keystore
			Pin string `json:"pin"`
		}

		data.Keystore = keystore
		data.Pin = pin

		m, err := json.Marshal(&data)
		if err != nil {
			cmd.PrintErrln(err)
			return
		}

		cmd.Println(string(m))
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.Flags().StringP("name", "n", "compound", "")
}
