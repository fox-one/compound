package cmd

import (
	"compound/worker"
	"compound/worker/market"
	"compound/worker/priceoracle"
	"compound/worker/snapshot"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "compound job worker",
	Run: func(cmd *cobra.Command, args []string) {
		// ctx := cmd.Context()

		db := provideDatabase()
		mainWallet := provideMainWallet()
		blockWallet := provideBlockWallet()
		config := provideConfig()

		propertyStore := providePropertyStore(db)
		marketStore := provideMarketStore()
		supplyStore := provideSupplyStore()
		borrowStore := provideBorrowStore()

		walletService := provideWalletService()
		blockService := provideBlockService()
		priceService := providePriceService()
		marketService := provideMarketService()
		supplyService := provideSupplyService()
		borrowService := provideBorrowService()
		accountService := provideAccountService()

		workers := []worker.IJob{
			priceoracle.New(mainWallet, blockWallet, config, marketStore, blockService, priceService),
			market.New(mainWallet, blockWallet, config, marketStore, blockService, priceService),
			snapshot.New(config,
				mainWallet,
				blockWallet,
				propertyStore,
				db,
				marketStore,
				supplyStore,
				borrowStore,
				walletService,
				priceService,
				blockService,
				marketService,
				supplyService,
				borrowService,
				accountService),
		}

		for _, w := range workers {
			w.Start()
			defer w.Stop()
		}

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, os.Kill)
		<-sig
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
