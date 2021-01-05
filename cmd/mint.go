package cmd

import (
	"compound/pkg/id"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var injectMintTokenCmd = &cobra.Command{
	Use:     "inject-ctoken",
	Aliases: []string{"ic"},
	Short:   "inject ctoken for mint",
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

		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		input := mixin.TransferInput{
			AssetID: assetID,
			Amount:  amountNum,
			TraceID: id.GenTraceID(),
			Memo:    "mint ctoken",
		}
		input.OpponentMultisig.Receivers = system.MemberIDs()
		input.OpponentMultisig.Threshold = system.Threshold

		payment, err := dapp.Client.VerifyPayment(ctx, input)
		if err != nil {
			panic(err)
		}

		url := mixin.URL.Codes(payment.CodeID)
		cmd.Println(url)
		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

func init() {
	rootCmd.AddCommand(injectMintTokenCmd)
	injectMintTokenCmd.Flags().StringP("asset", "a", "", "asset id")
	injectMintTokenCmd.Flags().StringP("amount", "q", "", "amount")
}
