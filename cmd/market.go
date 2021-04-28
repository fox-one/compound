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

// governing command for market
var addMarketCmd = &cobra.Command{
	Use:     "add-market",
	Aliases: []string{"am"},
	Short:   "Create a market",
	Long:    "s for symbol, a for asset_id, c for ctoken_asset_id",
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
		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalAddMarket), addMarketReq)
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

// governing command for market
var updateMarketCmd = &cobra.Command{
	Use:     "update-market",
	Aliases: []string{"um"},
	Short:   "update market parameters",
	Long:    "s for symbol, ie for init_exchange, rf for reserve_factor, li for liquidation_incentive, cf for collateral_factor, br for base_rate",
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

		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalUpdateMarket), updateMarketReq)
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

// governing command for market
var updateMarketAdvanceCmd = &cobra.Command{
	Use:     "update-market-advance",
	Aliases: []string{"uma"},
	Short:   "update market advance parameters",
	Long:    "s for symbol, bc for borrow_cap, clf for close_factor, m for multiplier, jm for jump_multiplier, k for kink",
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

		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalUpdateMarketAdvance), updateMarketReq)
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

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		closeMarketReq := proposal.MarketStatusReq{
			Status: core.MarketStatusClose,
		}

		asset, e := cmd.Flags().GetString("asset")
		if e != nil || asset == "" {
			panic("invalid asset")
		}
		closeMarketReq.AssetID = asset

		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalCloseMarket), closeMarketReq)
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

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		openMarketReq := proposal.MarketStatusReq{
			Status: core.MarketStatusOpen,
		}

		asset, e := cmd.Flags().GetString("asset")
		if e != nil || asset == "" {
			panic("invalid asset")
		}
		openMarketReq.AssetID = asset

		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalOpenMarket), openMarketReq)
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
	rootCmd.AddCommand(closeMarketCmd)
	rootCmd.AddCommand(openMarketCmd)

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

	closeMarketCmd.Flags().String("asset", "", "asset id")

	openMarketCmd.Flags().String("asset", "", "asset id")
}
