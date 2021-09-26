package cmd

import (
	"compound/core"
	"compound/core/proposal"

	"github.com/fox-one/pkg/qrcode"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

//
var withdrawCmd = &cobra.Command{
	Use:     "withdraw",
	Aliases: []string{"ww"},
	Short:   "Create a proposal for withdrawing from the mtg wallet",
	Long:    "opponent for receiver, asset for asset id, amount for withdrawing amount",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		opponent, err := cmd.Flags().GetString("opponent")
		if err != nil {
			panic(err)
		}

		asset, err := cmd.Flags().GetString("asset")
		if err != nil {
			panic(err)
		}

		amountStr, err := cmd.Flags().GetString("amount")
		if err != nil {
			panic(err)
		}
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			panic(err)
		}

		withdrawRequest := proposal.WithdrawReq{
			Opponent: opponent,
			Asset:    asset,
			Amount:   amount,
		}

		url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalWithdrawReserves, withdrawRequest)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		cmd.Println(url)
		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

func init() {
	proposalCmd.AddCommand(withdrawCmd)

	withdrawCmd.Flags().StringP("opponent", "o", "", "opponent id")
	withdrawCmd.Flags().StringP("asset", "s", "", "asset id")
	withdrawCmd.Flags().StringP("amount", "a", "", "asset amount")
}
