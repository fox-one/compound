package cmd

import (
	"compound/core"
	"context"
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

		blockService := provideBlockService()
		memo := make(core.BlockMemo)
		memo[core.BlockMemoKeyService] = core.MemoServiceMarket
		memo[core.BlockMemoKeyBlock] = "1234567890"
		memo[core.BlockMemoKeyUtilizationRate] = "0.5667"
		memo[core.BlockMemoKeyBorrowRate] = "0.0000000002324345"
		memo[core.BlockMemoKeySupplyRate] = "0.0000000003434535"

		fmt.Println(blockService.FormatBlockMemo(context.Background(), memo))
		fmt.Println(decimal.NewFromFloat(0.13).Div(decimal.NewFromInt(2102400)))
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
