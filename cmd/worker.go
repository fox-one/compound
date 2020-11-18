package cmd

import (
	"compound/worker"
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

		config := provideConfig()
		db := provideDatabase()
		mainWallet := provideMainWallet()
		blockWallet := provideBlockWallet()

		propertyStore := providePropertyStore(db)
		marketStore := provideMarketStore(db)
		supplyStore := provideSupplyStore(db)
		borrowStore := provideBorrowStore(db)

		walletService := provideWalletService(mainWallet)
		blockService := provideBlockService()
		priceService := providePriceService(blockService)
		marketService := provideMarketService(mainWallet, marketStore, borrowStore, blockService, priceService)
		accountService := provideAccountService(mainWallet, marketStore, supplyStore, borrowStore, priceService, blockService, walletService, marketService)
		supplyService := provideSupplyService(db, mainWallet, blockWallet, supplyStore, marketStore, accountService, priceService, blockService, walletService, marketService)
		borrowService := provideBorrowService(mainWallet, blockWallet, marketStore, borrowStore, blockService, priceService, walletService, accountService, marketService)

		workers := []worker.IJob{
			priceoracle.New(mainWallet, blockWallet, config, marketStore, blockService, priceService, walletService),
			snapshot.New(config, db, mainWallet, blockWallet, propertyStore, marketStore, supplyStore, borrowStore, walletService, priceService, blockService, marketService, supplyService, borrowService, accountService),
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
