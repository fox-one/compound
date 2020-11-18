package cmd

import (
	"compound/core"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		// now := time.Now()
		// l, _ := time.LoadLocation("Asia/Shanghai")
		// t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, l)
		// fmt.Println(t.UTC().Unix())
		// dappClient := provideMixinClient()
		// payment, e := dappClient.VerifyPayment(context.Background(), mixin.TransferInput{
		// 	AssetID:    "4d8c508b-91c5-375b-92b0-ee702ed2dac5",
		// 	OpponentID: "bfc9727a-87ca-4f0c-b105-a873f00eb53b",
		// 	Amount:     decimal.NewFromInt(1),
		// 	TraceID:    "d31034d1-2d28-4e2c-ac4d-ce5979dfd67b",
		// })
		// if e != nil {
		// 	cmd.PrintErrln(e)
		// 	return
		// }

		// bs, e := json.Marshal(payment)
		// if e != nil {
		// 	cmd.PrintErrln(e)
		// 	return
		// }

		// cmd.Println(string(bs))

		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceAddMarket
		memo[core.ActionKeySymbol] = "USDT"
		memo[core.ActionKeyBlock] = "1234567890"
		memo[core.ActionKeyBorrowRate] = "0.0000000002324345"
		memo[core.ActionKeySupplyRate] = "0.0000000003434535"

		mStr, _ := memo.Format()
		fmt.Println(mStr)
		fmt.Println(decimal.NewFromFloat(0.13).Div(decimal.NewFromInt(2102400)))

		// fmt.Println(wallet.PaySchemaURL(decimal.NewFromInt(12), "965e5c6e-434c-3fa9-b780-c50f43cd955c", "8be122b4-596f-4e4f-a307-978bed0ffb75", id.GenTraceID(), mStr))
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
