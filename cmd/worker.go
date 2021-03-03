package cmd

import (
	"compound/handler/hc"
	walletservice "compound/service/wallet"
	"compound/worker"
	"compound/worker/cashier"
	"compound/worker/message"
	"compound/worker/priceoracle"
	"compound/worker/snapshot"
	"compound/worker/spentsync"
	"compound/worker/syncer"
	"compound/worker/txsender"
	"fmt"
	"net/http"
	"sync"

	"github.com/fox-one/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "compound job worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		log := logger.FromContext(ctx)
		ctx = logger.WithContext(ctx, log)

		migrateDB()

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
		outputArchiveStore := provideOutputArchiveStore(db)
		allowListStore := provideAllowListStore(db)

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
		allowListService := provideAllowListService(propertyStore, allowListStore)

		//hc api
		{
			mux := chi.NewMux()
			mux.Use(middleware.Recoverer)
			mux.Use(middleware.StripSlashes)
			mux.Use(cors.AllowAll().Handler)
			mux.Use(logger.WithRequestID)
			mux.Use(middleware.Logger)

			mux.Mount("/hc", hc.Handle(rootCmd.Version))

			port, _ := cmd.Flags().GetInt("port")
			addr := fmt.Sprintf(":%d", port)

			go http.ListenAndServe(addr, mux)
		}

		workers := []worker.Worker{
			cashier.New(walletStore, walletService, system),
			message.New(messageStore, messageService),
			priceoracle.New(system, dapp, marketStore, priceStore, blockService, priceService),
			snapshot.NewPayee(db, system, dapp, propertyStore, userStore, outputArchiveStore, walletStore, priceStore, marketStore, supplyStore, borrowStore, proposalStore, transactionStore, proposalService, priceService, blockService, marketService, supplyService, borrowService, accountService, allowListService),
			syncer.New(walletStore, walletService, propertyStore),
			txsender.New(walletStore),
			spentsync.New(db, walletStore, transactionStore),
		}

		wg := sync.WaitGroup{}
		for _, w := range workers {
			wg.Add(1)

			go func(worker worker.Worker) {
				defer wg.Done()
				worker.Run(ctx)
			}(w)
		}

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.Flags().Int("port", 80, "worker api port")
}
