package cmd

import (
	"time"

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

		ticker, e := priceService.PullPriceTicker(ctx, "43d61dcd-e413-450d-80b8-101d5e903357", time.Time{})
		if e != nil {
			panic(e)
		}

		cmd.Println(ticker)
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
