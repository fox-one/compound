package cmd

import (
	"compound/pkg/id"
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
		// walletService := provideWalletService(provideMainWallet())
		// name, err := cmd.Flags().GetString("name")
		// if err != nil {
		// 	cmd.PrintErrln(err)
		// 	return
		// }

		// pin := mixin.RandomPin()
		// keystore, _, err := walletService.NewWallet(cmd.Context(), name, pin)
		// if err != nil {
		// 	cmd.PrintErrln(err)
		// 	return
		// }

		// var data struct {
		// 	*mixin.Keystore
		// 	Pin string `json:"pin"`
		// }

		// data.Keystore = keystore
		// data.Pin = pin

		// m, err := json.Marshal(&data)
		// if err != nil {
		// 	cmd.PrintErrln(err)
		// 	return
		// }

		// cmd.Println(string(m))
	},
}

var assetsCmd = &cobra.Command{
	Use:     "assets",
	Aliases: []string{"as"},
	Short:   "query wallet assets",
	Run: func(cmd *cobra.Command, args []string) {
		mainWallet := provideDapp()
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

var dappWithdrawCmd = &cobra.Command{
	Use:     "dwithdraw",
	Aliases: []string{"dw"},
	Short:   "query wallet assets",
	Run: func(cmd *cobra.Command, args []string) {
		mixin.UseApiHost(mixin.ZeromeshApiHost)
		dapp := provideDapp()
		cmd.Println("p:", dapp.Pin)
		cmd.Println(dapp.Client.ClientID)
		s, e := dapp.Client.Transfer(cmd.Context(), &mixin.TransferInput{
			AssetID:    "965e5c6e-434c-3fa9-b780-c50f43cd955c",
			OpponentID: "8be122b4-596f-4e4f-a307-978bed0ffb75",
			Amount:     decimal.NewFromFloat(1),
			TraceID:    id.GenTraceID(),
			Memo:       "w",
		}, dapp.Pin)

		if e != nil {
			panic(e)
		}

		fmt.Println(s)
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.Flags().StringP("name", "n", "compound", "")

	rootCmd.AddCommand(assetsCmd)
	rootCmd.AddCommand(dappWithdrawCmd)
}
