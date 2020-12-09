package cmd

import (
	"compound/core"
	"compound/pkg/id"
	"compound/pkg/mtg"
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var withdrawCmd = &cobra.Command{
	Use:     "withdraw",
	Aliases: []string{"ww"},
	Short:   "Create a proposal to withdraw from mtg wallet",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

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

		withdrawRequest := core.Withdraw{
			Opponent: opponent,
			Asset:    asset,
			Amount:   amount,
		}

		memo, err := mtg.Encode(clientID, traceID, int(core.ActionTypeProposalWithdraw), withdrawRequest)
		if err != nil {
			panic(err)
		}

		sign := mtg.Sign(memo, system.SignKey)
		memo = mtg.Pack(memo, sign)

		input := mixin.TransferInput{
			AssetID: system.VoteAsset,
			Amount:  system.VoteAmount,
			TraceID: traceID.String(),
			Memo:    base64.StdEncoding.EncodeToString(memo),
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
	proposalCmd.AddCommand(withdrawCmd)

	withdrawCmd.Flags().StringP("opponent", "o", "", "opponent id")
	withdrawCmd.Flags().StringP("asset", "s", "", "asset id")
	withdrawCmd.Flags().StringP("amount", "a", "", "asset amount")
}
