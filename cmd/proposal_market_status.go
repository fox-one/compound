package cmd

import (
	"compound/core"
	"compound/core/proposal"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/fox-one/pkg/uuid"
	"github.com/spf13/cobra"
)

// governing command for market
var closeMarketCmd = &cobra.Command{
	Use:     "close-market",
	Aliases: []string{"cm"},
	Short:   "close market",
	Long:    "close the market when it is under attack. if the market closed, tradings are disabled. asset for asset_id",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()
		exec, _ := cmd.Flags().GetBool("exec")

		assets, err := cmd.Flags().GetStringSlice("assets")
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		for _, asset := range assets {
			closeMarketReq := proposal.MarketStatusReq{
				Status:  core.MarketStatusClose,
				AssetID: asset,
			}
			if !exec || dapp.Pin == "" {
				url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalCloseMarket, closeMarketReq)
				if err != nil {
					cmd.PrintErr(err)
					return
				}

				cmd.Println(url)
				qrcode.Fprint(cmd.OutOrStdout(), url)
			} else {
				memo, err := buildProposalMemo(ctx, system, dapp.Client, core.ActionTypeProposalCloseMarket, closeMarketReq)
				if err != nil {
					cmd.PrintErr(err)
					return
				}
				input := &mixin.TransferInput{
					AssetID: system.VoteAsset,
					Amount:  system.VoteAmount,
					TraceID: uuid.New(),
					Memo:    memo,
				}
				input.OpponentMultisig.Threshold = system.Threshold
				input.OpponentMultisig.Receivers = system.MemberIDs
				if _, err := dapp.Client.Transaction(ctx, input, dapp.Pin); err != nil {
					cmd.PrintErr("transfer failed", err, asset)
					return
				}
				cmd.Println("close market request sent", asset)
			}
		}
	},
}

// governing command for market
var openMarketCmd = &cobra.Command{
	Use:     "open-market",
	Aliases: []string{"om"},
	Short:   "open market",
	Long:    "open the market when the attacking disapeared. asset for asset_id",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()
		exec, _ := cmd.Flags().GetBool("exec")

		assets, err := cmd.Flags().GetStringSlice("assets")
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		for _, asset := range assets {
			openMarketReq := proposal.MarketStatusReq{
				Status:  core.MarketStatusOpen,
				AssetID: asset,
			}

			if !exec || dapp.Pin == "" {
				url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalOpenMarket, openMarketReq)
				if err != nil {
					cmd.PrintErr(err)
					return
				}

				cmd.Println(url)
				qrcode.Fprint(cmd.OutOrStdout(), url)
			} else {
				memo, err := buildProposalMemo(ctx, system, dapp.Client, core.ActionTypeProposalOpenMarket, openMarketReq)
				if err != nil {
					cmd.PrintErr(err)
					return
				}
				input := &mixin.TransferInput{
					AssetID: system.VoteAsset,
					Amount:  system.VoteAmount,
					TraceID: uuid.New(),
					Memo:    memo,
				}
				input.OpponentMultisig.Threshold = system.Threshold
				input.OpponentMultisig.Receivers = system.MemberIDs
				if _, err := dapp.Client.Transaction(ctx, input, dapp.Pin); err != nil {
					cmd.PrintErr("transfer failed", err, asset)
					return
				}
				cmd.Println("open market request sent", asset)
			}
		}
	},
}

func init() {
	proposalCmd.AddCommand(closeMarketCmd)
	proposalCmd.AddCommand(openMarketCmd)

	closeMarketCmd.Flags().StringSlice("assets", nil, "asset id list")
	closeMarketCmd.Flags().Bool("exec", false, "exec upsert directly")

	openMarketCmd.Flags().StringSlice("assets", nil, "asset id list")
	openMarketCmd.Flags().Bool("exec", false, "exec upsert directly")
}
