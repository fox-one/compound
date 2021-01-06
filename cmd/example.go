package cmd

import (
	"compound/core"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		blockService := provideBlockService()
		priceService := providePriceService(blockService)

		t, e := priceService.PullPriceTicker(ctx, "c6d0c728-2624-429b-8e0d-d9d19b6592fa", time.Now())
		if e != nil {
			panic(e)
		}

		tbs, e := json.MarshalIndent(t, "", "    ")
		if e != nil {
			panic(e)
		}

		cmd.Println(string(tbs))
	},
}

func update(m *core.Market) {
	m.Price = decimal.NewFromInt(1200)
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
