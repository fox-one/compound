package cmd

import (
	walletservice "compound/service/wallet"
	"compound/worker"
	"compound/worker/cashier"
	"compound/worker/message"
	"compound/worker/priceoracle"
	"compound/worker/snapshot"
	"compound/worker/spentsync"
	"compound/worker/syncer"
	"compound/worker/txsender"

	"github.com/fox-one/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "compound job worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		log := logger.FromContext(ctx)
		ctx = logger.WithContext(ctx, log)

		db := provideDatabase()
		dapp := provideDapp()
		system := provideSystem()

		propertyStore := providePropertyStore(db)
		marketStore := provideMarketStore(db)
		supplyStore := provideSupplyStore(db)
		borrowStore := provideBorrowStore(db)
		walletStore := provideWalletStore(db)
		messageStore := provideMessageStore(db)
		priceStore := providePriceStore(db)
		proposalStore := provideProposalStore(db)
		userStore := provideUserStore(db)
		transactionStore := provideTransactionStore(db)

		walletService := provideWalletService(dapp.Client, walletservice.Config{
			Pin:       dapp.Pin,
			Members:   system.MemberIDs(),
			Threshold: system.Threshold,
		})

		blockService := provideBlockService()
		priceService := providePriceService(blockService)
		marketService := provideMarketService(marketStore, blockService)
		accountService := provideAccountService(marketStore, supplyStore, borrowStore, priceService, blockService, marketService)
		supplyService := provideSupplyService(marketService)
		borrowService := provideBorrowService(blockService, priceService, accountService)
		messageService := provideMessageService(dapp.Client)
		proposalService := provideProposalService(dapp.Client, system, marketStore, messageStore)

		workers := []worker.Worker{
			cashier.New(walletStore, walletService, system),
			message.New(messageStore, messageService),
			priceoracle.New(system, dapp, marketStore, priceStore, blockService, priceService),
			snapshot.NewPayee(db, system, dapp, propertyStore, userStore, walletStore, priceStore, marketStore, supplyStore, borrowStore, proposalStore, transactionStore, proposalService, priceService, blockService, marketService, supplyService, borrowService, accountService),
			syncer.New(walletStore, walletService, propertyStore),
			txsender.New(walletStore),
			spentsync.New(db, walletStore, transactionStore),
		}

		var g errgroup.Group
		for _, w := range workers {
			worker := w
			g.Go(func() error {
				return worker.Run(ctx)
			})
		}

		if err := g.Wait(); err != nil {
			cmd.PrintErrln("run worker error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
