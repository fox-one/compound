package cmd

import (
	"compound/core"
	"compound/pkg/id"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var addMarketCmd = &cobra.Command{
	Use:     "add-market",
	Aliases: []string{"am"},
	Short:   "add market",
	Run: func(cmd *cobra.Command, args []string) {
		mainWallet := provideMainWallet()
		walletService := provideWalletService(mainWallet)
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

		amount := decimal.NewFromFloat(0.00000001)
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceAddMarket
		action[core.ActionKeySymbol] = strings.ToUpper(symbol)
		action[core.ActionKeyAssetID] = assetID
		action[core.ActionKeyCTokenAssetID] = ctokenAssetID
		memoStr, e := action.Format()
		if e != nil {
			panic(e)
		}

		url, e := walletService.PaySchemaURL(amount, cfg.App.GasAssetID, mainWallet.Client.ClientID, id.GenTraceID(), memoStr)
		if e != nil {
			panic(e)
		}

		fmt.Println(url)
	},
}

var updateMarketCmd = &cobra.Command{
	Use:     "update-market",
	Aliases: []string{"um"},
	Short:   "update market",
	Run: func(cmd *cobra.Command, args []string) {
		mainWallet := provideMainWallet()
		walletService := provideWalletService(mainWallet)

		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceUpdateMarket
		symbol, e := cmd.Flags().GetString("s")
		if e != nil || symbol == "" {
			panic("invalid symbol")
		}
		action[core.ActionKeySymbol] = strings.ToUpper(symbol)

		flag, e := cmd.Flags().GetString("ie")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyInitExchangeRate] = flag

		flag, e = cmd.Flags().GetString("rf")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyReserveFactor] = flag

		flag, e = cmd.Flags().GetString("li")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyLiquidationIncentive] = flag

		flag, e = cmd.Flags().GetString("bc")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyBorrowCap] = flag

		flag, e = cmd.Flags().GetString("cf")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyCollateralFactor] = flag

		flag, e = cmd.Flags().GetString("clf")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyCloseFactor] = flag

		flag, e = cmd.Flags().GetString("br")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyBaseRate] = flag

		flag, e = cmd.Flags().GetString("m")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyMultiPlier] = flag

		flag, e = cmd.Flags().GetString("jm")
		if e != nil {
			panic("invalid flag")
		}

		action[core.ActionKeyJumpMultiPlier] = flag

		flag, e = cmd.Flags().GetString("ie")
		if e != nil {
			panic("invalid flag")
		}
		action[core.ActionKeyKink] = flag

		memoStr, e := action.Format()
		if e != nil {
			panic(e)
		}

		url, e := walletService.PaySchemaURL(decimal.NewFromFloat(0.00000001), cfg.App.GasAssetID, mainWallet.Client.ClientID, id.GenTraceID(), memoStr)
		if e != nil {
			panic(e)
		}

		fmt.Println(url)
	},
}

func init() {
	rootCmd.AddCommand(addMarketCmd)
	rootCmd.AddCommand(updateMarketCmd)

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
}
