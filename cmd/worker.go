package cmd

import (
	"compound/worker"
	"compound/worker/priceoracle"
	"compound/worker/snapshot"
	"compound/worker/transfer"
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
		transferStore := provideTransferStore(db)

		walletService := provideWalletService(mainWallet)
		blockService := provideBlockService()
		priceService := providePriceService(blockService)
		marketService := provideMarketService(mainWallet, marketStore, borrowStore, blockService, priceService)
		accountService := provideAccountService(mainWallet, marketStore, supplyStore, borrowStore, priceService, blockService, walletService, marketService)
		supplyService := provideSupplyService(db, mainWallet, blockWallet, supplyStore, marketStore, accountService, priceService, blockService, walletService, marketService)
		borrowService := provideBorrowService(mainWallet, blockWallet, marketStore, borrowStore, blockService, priceService, walletService, accountService, marketService)

		workers := []worker.IJob{
			priceoracle.New(mainWallet, blockWallet, config, marketStore, blockService, priceService, walletService),
			transfer.New(db, mainWallet, config, transferStore, walletService),
			snapshot.New(config, db, mainWallet, blockWallet, propertyStore, transferStore, marketStore, supplyStore, borrowStore, walletService, priceService, blockService, marketService, supplyService, borrowService, accountService),
		}

		for _, w := range workers {
			w.Start()
		}

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, os.Kill)
		<-sig
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
