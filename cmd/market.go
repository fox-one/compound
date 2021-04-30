package cmd

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/id"
	"compound/pkg/mtg"
	"encoding/base64"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

// governing command for market
var addMarketCmd = &cobra.Command{
	Use:     "market",
	Aliases: []string{"m"},
	Short:   "Create or update market",
	Long: `params->
	symbol: symbol name
	asset: underlying asset id
	ctoken: ctoken asset id
	init_exchange_rate: init exchange rate
	reserve_factor: reserve factor
	liquidation_incentive: liquidation incentive
	borrow_cap: borrow cap
	collateral_factor: collateral factor
	close_factor: close factor
	base_rate: base rate
	multi: multiplier
	jump_multi: jump multiplier
	kink: kink`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		req := proposal.MarketReq{}

		symbol, e := cmd.Flags().GetString("symbol")
		if e != nil || symbol == "" {
			panic("invalid symbol")
		}
		req.Symbol = symbol

		assetID, e := cmd.Flags().GetString("asset")
		if e != nil || assetID == "" {
			panic("invalid assetID")
		}
		req.AssetID = assetID

		ctokenAssetID, e := cmd.Flags().GetString("ctoken")
		if e != nil || ctokenAssetID == "" {
			panic("invalid ctokenAssetID")
		}
		req.CTokenAssetID = ctokenAssetID

		flag, e := cmd.Flags().GetString("init_exchange_rate")
		if e != nil {
			panic("invalid flag")
		}
		ie, _ := decimal.NewFromString(flag)
		req.InitExchange = ie

		flag, e = cmd.Flags().GetString("reserve_factor")
		if e != nil {
			panic("invalid flag")
		}
		rf, _ := decimal.NewFromString(flag)
		req.ReserveFactor = rf

		flag, e = cmd.Flags().GetString("liquidation_incentive")
		if e != nil {
			panic("invalid flag")
		}
		li, _ := decimal.NewFromString(flag)
		req.LiquidationIncentive = li

		flag, e = cmd.Flags().GetString("collateral_factor")
		if e != nil {
			panic("invalid flag")
		}
		cf, _ := decimal.NewFromString(flag)
		req.CollateralFactor = cf

		flag, e = cmd.Flags().GetString("base_rate")
		if e != nil {
			panic("invalid flag")
		}
		br, _ := decimal.NewFromString(flag)
		req.BaseRate = br

		flag, e = cmd.Flags().GetString("borrow_cap")
		if e != nil {
			panic("invalid flag")
		}
		bc, _ := decimal.NewFromString(flag)
		req.BorrowCap = bc

		flag, e = cmd.Flags().GetString("close_factor")
		if e != nil {
			panic("invalid flag")
		}
		clf, _ := decimal.NewFromString(flag)
		req.CloseFactor = clf

		flag, e = cmd.Flags().GetString("multi")
		if e != nil {
			panic("invalid flag")
		}
		m, _ := decimal.NewFromString(flag)
		req.Multiplier = m

		flag, e = cmd.Flags().GetString("jump_multi")
		if e != nil {
			panic("invalid flag")
		}

		jm, _ := decimal.NewFromString(flag)
		req.JumpMultiplier = jm

		flag, e = cmd.Flags().GetString("kink")
		if e != nil {
			panic("invalid flag")
		}
		k, _ := decimal.NewFromString(flag)
		req.Kink = k

		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalAddMarket), req)
		if err != nil {
			panic(err)
		}

		sign := mtg.Sign(memo, system.SignKey)
		signedMemo := mtg.Pack(memo, sign)

		memoStr := base64.StdEncoding.EncodeToString(signedMemo)
		fmt.Println("memo.length:", len(memoStr))

		if len(memoStr) > 200 {
			memoStr = base64.StdEncoding.EncodeToString(memo)
		}
		fmt.Println("memo.length:", len(memoStr))

		input := mixin.TransferInput{
			AssetID: system.VoteAsset,
			Amount:  system.VoteAmount,
			TraceID: traceID.String(),
			Memo:    memoStr,
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
	rootCmd.AddCommand(closeMarketCmd)
	rootCmd.AddCommand(openMarketCmd)

	addMarketCmd.Flags().String("symbol", "", "market symbol")
	addMarketCmd.Flags().String("asset", "", "asset id")
	addMarketCmd.Flags().String("ctoken", "", "ctoken asset id")
	addMarketCmd.Flags().String("init_exchange_rate", "", "intial exchange rate")
	addMarketCmd.Flags().String("reserve_factor", "", "reserve factor")
	addMarketCmd.Flags().String("liquidation_incentive", "", "liquidation incentive")
	addMarketCmd.Flags().String("borrow_cap", "", "borrow cap")
	addMarketCmd.Flags().String("collateral_factor", "", "collateral factor")
	addMarketCmd.Flags().String("close_factor", "", "close factor")
	addMarketCmd.Flags().String("base_rate", "", "base rate")
	addMarketCmd.Flags().String("multi", "", "multiplier")
	addMarketCmd.Flags().String("jump_multi", "", "jump multiplier")
	addMarketCmd.Flags().String("kink", "", "kink")

	closeMarketCmd.Flags().String("asset", "", "asset id")

	openMarketCmd.Flags().String("asset", "", "asset id")
}
