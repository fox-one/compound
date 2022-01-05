package cmd

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/number"
	"encoding/csv"
	"os"
	"strconv"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/fox-one/pkg/uuid"
	"github.com/spf13/cobra"
)

// governing command for upsert markets
var upsertMarketsCmd = &cobra.Command{
	Use: "markets",
	Long: "input csv file with the following format:\n" +
		"Symbol,Asset,C Token,Init Exchange,Reserve Factor,Liquidation Incentive,Collateral Factor,Base Rate,Close Factor,Multiplier,Jump Multiplier,Kink,Price Threshold,Price,Submit State,Duplicated",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()
		exec, _ := cmd.Flags().GetBool("exec")

		csvFile, err := cmd.Flags().GetString("input")
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		file, err := os.Open(csvFile)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		rows, err := csv.NewReader(file).ReadAll()
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		for _, row := range rows {
			if len(row) < 13 {
				cmd.PrintErr("invalid csv format", row)
				return
			}

			var req = &proposal.MarketReq{
				Symbol:               row[0],
				AssetID:              row[1],
				CTokenAssetID:        row[2],
				InitExchange:         number.Decimal(row[3]),
				ReserveFactor:        number.Decimal(row[4]),
				LiquidationIncentive: number.Decimal(row[5]),
				CollateralFactor:     number.Decimal(row[6]),
				BaseRate:             number.Decimal(row[7]),
				CloseFactor:          number.Decimal(row[8]),
				Multiplier:           number.Decimal(row[9]),
				JumpMultiplier:       number.Decimal(row[10]),
				Kink:                 number.Decimal(row[11]),
			}
			threshold, err := strconv.ParseInt(row[12], 10, 64)
			if err != nil {
				cmd.PrintErr("invalid price threshold", row, row[12])
				return
			}
			req.PriceThreshold = int(threshold)

			asset, err := dapp.Client.ReadAsset(ctx, req.AssetID)
			if err != nil {
				cmd.PrintErr("read asset failed", err, req.AssetID)
				return
			}

			if asset.Symbol != req.Symbol {
				cmd.PrintErr("symbol not matched", req.Symbol, asset.Symbol)
				return
			}
			req.Price = asset.PriceUSD

			if !exec || dapp.Pin == "" {
				url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalUpsertMarket, req)
				if err != nil {
					cmd.PrintErr(err)
					return
				}
				cmd.Println(req.Symbol, req.AssetID, url)
				qrcode.Fprint(cmd.OutOrStdout(), url)
			} else {
				memo, err := buildProposalMemo(ctx, system, dapp.Client, core.ActionTypeProposalUpsertMarket, req)
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
					cmd.PrintErr("transfer failed", err, req.AssetID)
					return
				}
				cmd.Println("upsert market request sent", req.Symbol)
			}
		}
	},
}

func init() {
	proposalCmd.AddCommand(upsertMarketsCmd)
	upsertMarketsCmd.Flags().String("input", "assets.csv", "input assets csv")
	upsertMarketsCmd.Flags().Bool("exec", false, "exec upsert directly")
}
