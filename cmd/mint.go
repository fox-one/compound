package cmd

import (
	"compound/core"
	"compound/pkg/id"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var injectMintTokenCmd = &cobra.Command{
	Use:     "inject-mint-token",
	Aliases: []string{"imt"},
	Short:   "inject mint token",
	Run: func(cmd *cobra.Command, args []string) {
		assetID, e := cmd.Flags().GetString("asset")
		if e != nil || assetID == "" {
			panic(e)
		}

		amount, e := cmd.Flags().GetString("amount")
		if e != nil {
			panic(e)
		}
		amountNum, e := decimal.NewFromString(amount)
		if e != nil || amountNum.LessThanOrEqual(decimal.Zero) {
			panic("invalid amount")
		}

		mainWallet := provideMainWallet()
		walletService := provideWalletService(mainWallet)

		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceInjectMintToken
		memoStr, e := action.Format()
		if e != nil {
			panic(e)
		}

		url, e := walletService.PaySchemaURL(amountNum, assetID, mainWallet.Client.ClientID, id.GenTraceID(), memoStr)
		if e != nil {
			panic(e)
		}

		fmt.Println(url)
	},
}

func init() {
	rootCmd.AddCommand(injectMintTokenCmd)
	injectMintTokenCmd.Flags().StringP("asset", "a", "", "asset id")
	injectMintTokenCmd.Flags().StringP("amount", "q", "", "amount")
}
