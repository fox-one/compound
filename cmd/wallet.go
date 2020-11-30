package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var walletCmd = &cobra.Command{
	Use:     "new-wallet",
	Aliases: []string{"nw"},
	Short:   "new wallet",
	Run: func(cmd *cobra.Command, args []string) {
		walletService := provideWalletService(provideMainWallet())
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

var assetsCmd = &cobra.Command{
	Use:     "assets",
	Aliases: []string{"as"},
	Short:   "query wallet assets",
	Run: func(cmd *cobra.Command, args []string) {
		mainWallet := provideMainWallet()
		assets, e := mainWallet.Client.ReadAssets(cmd.Context())
		if e != nil {
			panic(e)
		}

		result := make([]*mixin.Asset, 0)
		for _, a := range assets {
			if a.Balance.GreaterThan(decimal.Zero) {
				result = append(result, a)
			}
		}

		abs, e := json.Marshal(result)
		if e != nil {
			panic(e)
		}

		fmt.Println(string(abs))
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.Flags().StringP("name", "n", "compound", "")

	rootCmd.AddCommand(assetsCmd)
}
