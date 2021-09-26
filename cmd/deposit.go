package cmd

import (
	"compound/pkg/id"
	"encoding/json"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

// governing command for market
var depositCmd = &cobra.Command{
	Use:     "deposit",
	Aliases: []string{"ic"},
	Short:   "deposit",
	Long:    "deposit into the market before trading. asset for asset_id, amount for ctoken_amount",
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
		if e != nil || !amountNum.IsPositive() {
			panic("invalid amount")
		}

		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		input := mixin.TransferInput{
			AssetID: assetID,
			Amount:  amountNum,
			TraceID: id.GenTraceID(),
			Memo:    "deposit",
		}
		input.OpponentMultisig.Receivers = system.MemberIDs
		input.OpponentMultisig.Threshold = system.Threshold

		payment, err := dapp.Client.VerifyPayment(ctx, input)
		if err != nil {
			panic(err)
		}

		ibs, err := json.MarshalIndent(input, "", "    ")
		if err != nil {
			panic(err)
		}

		cmd.Println(string(ibs))

		url := mixin.URL.Codes(payment.CodeID)
		cmd.Println(url)
		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

func init() {
	rootCmd.AddCommand(depositCmd)
	depositCmd.Flags().StringP("asset", "a", "", "asset id")
	depositCmd.Flags().StringP("amount", "q", "", "amount")
}
