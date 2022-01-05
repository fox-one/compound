package cmd

import (
	"compound/core"
	"compound/core/proposal"

	"github.com/fox-one/pkg/qrcode"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

// governing command for market
var upsertMarketCmd = &cobra.Command{
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

		url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalUpsertMarket, req)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		cmd.Println(url)
		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

func init() {
	proposalCmd.AddCommand(upsertMarketCmd)

	upsertMarketCmd.Flags().String("symbol", "", "market symbol")
	upsertMarketCmd.Flags().String("asset", "", "asset id")
	upsertMarketCmd.Flags().String("ctoken", "", "ctoken asset id")
	upsertMarketCmd.Flags().String("init_exchange_rate", "1", "intial exchange rate")
	upsertMarketCmd.Flags().String("reserve_factor", "0.1", "reserve factor")
	upsertMarketCmd.Flags().String("liquidation_incentive", "0.05", "liquidation incentive")
	upsertMarketCmd.Flags().String("borrow_cap", "0", "borrow cap")
	upsertMarketCmd.Flags().String("collateral_factor", "0.75", "collateral factor")
	upsertMarketCmd.Flags().String("close_factor", "0.5", "close factor")
	upsertMarketCmd.Flags().String("base_rate", "0.025", "base rate")
	upsertMarketCmd.Flags().String("multi", "0.4", "multiplier")
	upsertMarketCmd.Flags().String("jump_multi", "0", "jump multiplier")
	upsertMarketCmd.Flags().String("kink", "0", "kink")
	upsertMarketCmd.Flags().String("price", "0", "price")
	upsertMarketCmd.Flags().Int("price_threshold", 0, "price threshold")
}
