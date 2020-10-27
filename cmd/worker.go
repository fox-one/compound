package cmd

import (
	"compound/worker"
	"compound/worker/snapshot"
	"compound/worker/block"
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
		dapp := provideMixinClient()
		blockWallet := provideBlockWallet()
		config := provideConfig()
		propertyStore := providePropertyStore(db)
		walletService := provideWalletService()
		blockService := provideBlockService()

		workers := []worker.IJob{
			block.New(config, dapp, blockWallet, blockService),
			snapshot.New(config, dapp, propertyStore, walletService, blockService),
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
