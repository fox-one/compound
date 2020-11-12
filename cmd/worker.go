package cmd

import (
	"compound/worker"
	"compound/worker/interest"
	"compound/worker/liquidity"
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

		config := provideConfig()
		db := provideDatabase()
		redis := provideRedis()
		mainWallet := provideMainWallet()
		blockWallet := provideBlockWallet()

		propertyStore := providePropertyStore(db)
		marketStore := provideMarketStore(db)
		supplyStore := provideSupplyStore(db)
		borrowStore := provideBorrowStore(db)
		accountStore := provideAccountStore(redis)

		walletService := provideWalletService(mainWallet)
		blockService := provideBlockService()
		priceService := providePriceService(redis, blockService)
		marketService := provideMarketService(redis, mainWallet, marketStore, borrowStore, blockService, priceService)
		accountService := provideAccountService(mainWallet, marketStore, supplyStore, borrowStore, accountStore, priceService, blockService, walletService, marketService)
		supplyService := provideSupplyService(db, mainWallet, blockWallet, supplyStore, marketStore, accountService, priceService, blockService, walletService, marketService)
		borrowService := provideBorrowService(mainWallet, blockWallet, marketStore, borrowStore, blockService, priceService, walletService, accountService)

		workers := []worker.IJob{
			priceoracle.New(mainWallet, blockWallet, config, marketStore, blockService, priceService),
			market.New(mainWallet, blockWallet, config, marketStore, blockService, priceService),
			interest.New(config, mainWallet, blockWallet, marketStore, supplyStore, borrowStore, blockService, marketService, walletService),
			liquidity.New(config, mainWallet, blockWallet, marketStore, supplyStore, borrowStore, blockService, marketService, walletService, accountService),
			snapshot.New(config, db, mainWallet, blockWallet, propertyStore, marketStore, supplyStore, borrowStore, accountStore, walletService, priceService, blockService, marketService, supplyService, borrowService, accountService),
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
