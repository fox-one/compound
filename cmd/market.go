package cmd

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/id"
	"compound/pkg/mtg"
	"encoding/base64"
	"strings"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var addMarketCmd = &cobra.Command{
	Use:     "add-market",
	Aliases: []string{"am"},
	Short:   "add market",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		symbol, e := cmd.Flags().GetString("s")
		if e != nil || symbol == "" {
			panic("invalid symbol")
		}
		assetID, e := cmd.Flags().GetString("a")
		if e != nil || assetID == "" {
			panic("invalid assetID")
		}
		ctokenAssetID, e := cmd.Flags().GetString("c")
		if e != nil || ctokenAssetID == "" {
			panic("invalid ctokenAssetID")
		}

		addMarketReq := proposal.AddMarketReq{
			Symbol:        strings.ToUpper(symbol),
			AssetID:       assetID,
			CTokenAssetID: ctokenAssetID,
		}
		memo, err := mtg.Encode(clientID, traceID, int(core.ActionTypeProposalAddMarket), addMarketReq)
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

var updateMarketCmd = &cobra.Command{
	Use:     "update-market",
	Aliases: []string{"um"},
	Short:   "update market",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		updateMarketReq := proposal.UpdateMarketReq{}

		symbol, e := cmd.Flags().GetString("s")
		if e != nil || symbol == "" {
			panic("invalid symbol")
		}
		updateMarketReq.Symbol = strings.ToUpper(symbol)

		flag, e := cmd.Flags().GetString("ie")
		if e != nil {
			panic("invalid flag")
		}
		ie, _ := decimal.NewFromString(flag)
		updateMarketReq.InitExchange = ie

		flag, e = cmd.Flags().GetString("rf")
		if e != nil {
			panic("invalid flag")
		}
		rf, _ := decimal.NewFromString(flag)
		updateMarketReq.ReserveFactor = rf

		flag, e = cmd.Flags().GetString("li")
		if e != nil {
			panic("invalid flag")
		}
		li, _ := decimal.NewFromString(flag)
		updateMarketReq.LiquidationIncentive = li

		flag, e = cmd.Flags().GetString("cf")
		if e != nil {
			panic("invalid flag")
		}
		cf, _ := decimal.NewFromString(flag)
		updateMarketReq.CollateralFactor = cf

		flag, e = cmd.Flags().GetString("br")
		if e != nil {
			panic("invalid flag")
		}
		br, _ := decimal.NewFromString(flag)
		updateMarketReq.BaseRate = br

		memo, err := mtg.Encode(clientID, traceID, int(core.ActionTypeProposalUpdateMarket), updateMarketReq)
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

var updateMarketAdvanceCmd = &cobra.Command{
	Use:     "update-market-advance",
	Aliases: []string{"uma"},
	Short:   "update market advance",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		updateMarketReq := proposal.UpdateMarketAdvanceReq{}

		symbol, e := cmd.Flags().GetString("s")
		if e != nil || symbol == "" {
			panic("invalid symbol")
		}
		updateMarketReq.Symbol = strings.ToUpper(symbol)

		flag, e := cmd.Flags().GetString("bc")
		if e != nil {
			panic("invalid flag")
		}
		bc, _ := decimal.NewFromString(flag)
		updateMarketReq.BorrowCap = bc

		flag, e = cmd.Flags().GetString("clf")
		if e != nil {
			panic("invalid flag")
		}
		clf, _ := decimal.NewFromString(flag)
		updateMarketReq.CloseFactor = clf

		flag, e = cmd.Flags().GetString("m")
		if e != nil {
			panic("invalid flag")
		}
		m, _ := decimal.NewFromString(flag)
		updateMarketReq.Multiplier = m

		flag, e = cmd.Flags().GetString("jm")
		if e != nil {
			panic("invalid flag")
		}

		jm, _ := decimal.NewFromString(flag)
		updateMarketReq.JumpMultiplier = jm

		flag, e = cmd.Flags().GetString("k")
		if e != nil {
			panic("invalid flag")
		}
		k, _ := decimal.NewFromString(flag)
		updateMarketReq.Kink = k

		memo, err := mtg.Encode(clientID, traceID, int(core.ActionTypeProposalUpdateMarketAdvance), updateMarketReq)
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
	rootCmd.AddCommand(addMarketCmd)
	rootCmd.AddCommand(updateMarketCmd)
	rootCmd.AddCommand(updateMarketAdvanceCmd)

	addMarketCmd.Flags().String("s", "", "market symbol")
	addMarketCmd.Flags().String("a", "", "asset id")
	addMarketCmd.Flags().String("c", "", "ctoken asset id")

	updateMarketCmd.Flags().String("s", "", "market symbol")
	updateMarketCmd.Flags().String("ie", "", "intial exchange rate")
	updateMarketCmd.Flags().String("rf", "", "reserve factor")
	updateMarketCmd.Flags().String("li", "", "liquidation incentive")
	updateMarketCmd.Flags().String("bc", "", "borrow cap")
	updateMarketCmd.Flags().String("cf", "", "collateral factor")
	updateMarketCmd.Flags().String("clf", "", "close factor")
	updateMarketCmd.Flags().String("br", "", "base rate")
	updateMarketCmd.Flags().String("m", "", "multiplier")
	updateMarketCmd.Flags().String("jm", "", "jump multiplier")
	updateMarketCmd.Flags().String("k", "", "kink")

	updateMarketAdvanceCmd.Flags().String("s", "", "market symbol")
	updateMarketAdvanceCmd.Flags().String("ie", "", "intial exchange rate")
	updateMarketAdvanceCmd.Flags().String("rf", "", "reserve factor")
	updateMarketAdvanceCmd.Flags().String("li", "", "liquidation incentive")
	updateMarketAdvanceCmd.Flags().String("bc", "", "borrow cap")
	updateMarketAdvanceCmd.Flags().String("cf", "", "collateral factor")
	updateMarketAdvanceCmd.Flags().String("clf", "", "close factor")
	updateMarketAdvanceCmd.Flags().String("br", "", "base rate")
	updateMarketAdvanceCmd.Flags().String("m", "", "multiplier")
	updateMarketAdvanceCmd.Flags().String("jm", "", "jump multiplier")
	updateMarketAdvanceCmd.Flags().String("k", "", "kink")
}
