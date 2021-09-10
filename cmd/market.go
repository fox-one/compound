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
	kink: kink
	price_threshold: int`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, err := uuid.FromString(system.ClientID)
		if err != nil {
			panic(err)
		}
		traceID, err := uuid.FromString(id.GenTraceID())
		if err != nil {
			panic(err)
		}

		req := proposal.MarketReq{}

		symbol, e := cmd.Flags().GetString("symbol")
		if e != nil {
			panic("invalid symbol")
		}
		req.Symbol = symbol

		assetID, e := cmd.Flags().GetString("asset")
		if e != nil || assetID == "" {
			panic("invalid assetID")
		}
		req.AssetID = assetID

		ctokenAssetID, e := cmd.Flags().GetString("ctoken")
		if e != nil {
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
		rf, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.ReserveFactor = rf

		flag, e = cmd.Flags().GetString("liquidation_incentive")
		if e != nil {
			panic("invalid flag")
		}
		li, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.LiquidationIncentive = li

		flag, e = cmd.Flags().GetString("collateral_factor")
		if e != nil {
			panic("invalid flag")
		}
		cf, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.CollateralFactor = cf

		flag, e = cmd.Flags().GetString("base_rate")
		if e != nil {
			panic("invalid flag")
		}
		br, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.BaseRate = br

		flag, e = cmd.Flags().GetString("borrow_cap")
		if e != nil {
			panic("invalid flag")
		}
		bc, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.BorrowCap = bc

		flag, e = cmd.Flags().GetString("close_factor")
		if e != nil {
			panic("invalid flag")
		}
		clf, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.CloseFactor = clf

		flag, e = cmd.Flags().GetString("multi")
		if e != nil {
			panic("invalid flag")
		}
		m, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.Multiplier = m

		flag, e = cmd.Flags().GetString("jump_multi")
		if e != nil {
			panic("invalid flag")
		}

		jm, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.JumpMultiplier = jm

		flag, e = cmd.Flags().GetString("kink")
		if e != nil {
			panic("invalid flag")
		}
		k, e := decimal.NewFromString(flag)
		if e != nil {
			panic(e)
		}
		req.Kink = k

		if pt, err := cmd.Flags().GetInt("price_threshold"); err != nil {
			panic("invalid param: price_threshold")
		} else {
			req.PriceThreshold = pt
		}

		{
			if price, err := cmd.Flags().GetString("price"); err != nil {
				panic("invalid price")
			} else if value, err := decimal.NewFromString(price); err != nil {
				panic(err)
			} else {
				req.Price = value
			}
		}

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

		clientID, err := uuid.FromString(system.ClientID)
		if err != nil {
			panic(err)
		}
		traceID, err := uuid.FromString(id.GenTraceID())
		if err != nil {
			panic(err)
		}

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

		clientID, err := uuid.FromString(system.ClientID)
		if err != nil {
			panic(err)
		}
		traceID, err := uuid.FromString(id.GenTraceID())
		if err != nil {
			panic(err)
		}

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
	addMarketCmd.Flags().String("init_exchange_rate", "1", "intial exchange rate")
	addMarketCmd.Flags().String("reserve_factor", "0.1", "reserve factor")
	addMarketCmd.Flags().String("liquidation_incentive", "0.05", "liquidation incentive")
	addMarketCmd.Flags().String("borrow_cap", "0", "borrow cap")
	addMarketCmd.Flags().String("collateral_factor", "0.75", "collateral factor")
	addMarketCmd.Flags().String("close_factor", "0.5", "close factor")
	addMarketCmd.Flags().String("base_rate", "0.025", "base rate")
	addMarketCmd.Flags().String("multi", "0.4", "multiplier")
	addMarketCmd.Flags().String("jump_multi", "0", "jump multiplier")
	addMarketCmd.Flags().String("kink", "0", "kink")
	addMarketCmd.Flags().String("price", "0", "price")
	addMarketCmd.Flags().Int("price_threshold", 0, "price threshold")

	closeMarketCmd.Flags().String("asset", "", "asset id")

	openMarketCmd.Flags().String("asset", "", "asset id")
}
