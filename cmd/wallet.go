package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

//获取dapp资产情况
var assetsCmd = &cobra.Command{
	Use:     "assets",
	Aliases: []string{"as"},
	Short:   "query wallet assets",
	Run: func(cmd *cobra.Command, args []string) {
		mainWallet := provideDapp()
		assets, e := mainWallet.Client.ReadAssets(cmd.Context())
		if e != nil {
			panic(e)
		}

		result := make([]*mixin.Asset, 0)
		for _, a := range assets {
			if a.Balance.GreaterThan(decimal.Zero) {
				result = append(result, a)
			}
		}

		abs, e := json.Marshal(result)
		if e != nil {
			panic(e)
		}

		fmt.Println(string(abs))
	},
}

func init() {
	rootCmd.AddCommand(assetsCmd)
}
