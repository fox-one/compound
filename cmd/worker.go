package cmd

import (
	"compound/handler/hc"
	walletservice "compound/service/wallet"
	"compound/worker"
	"compound/worker/assigner"
	"compound/worker/cashier"
	"compound/worker/datadog"
	"compound/worker/message"
	"compound/worker/payee"
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

// command for background worker
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "compound job worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		log := logger.FromContext(ctx)
		ctx = logger.WithContext(ctx, log)

		db := provideDatabase()
		defer db.Close()

		dapp := provideDapp()
		system := provideSystem()

		propertyStore := providePropertyStore(db)
		marketStore := provideMarketStore(db)
		supplyStore := provideSupplyStore(db)
		borrowStore := provideBorrowStore(db)
		walletStore := provideWalletStore(db)
		messageStore := provideMessageStore(db)
		proposalStore := provideProposalStore(db)
		userStore := provideUserStore(db)
		transactionStore := provideTransactionStore(db)
		allowListStore := provideAllowListStore(db)
		oracleSignerStore := provideOracleSignerStore(db)

		walletService := provideWalletService(dapp.Client, walletservice.Config{
			Pin:       dapp.Pin,
			Members:   system.MemberIDs(),
			Threshold: system.Threshold,
		})

		blockService := provideBlockService()
		marketService := provideMarketService(blockService)
		accountService := provideAccountService(marketStore, supplyStore, borrowStore, blockService, marketService)
		supplyService := provideSupplyService(marketService)
		borrowService := provideBorrowService(blockService, accountService)
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

			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				panic(err)
			}
			addr := fmt.Sprintf(":%d", port)

			go http.ListenAndServe(addr, mux)
		}

		workers := []worker.Worker{
			cashier.New(walletStore, walletService, system, provideCashierConfig()),
			assigner.New(walletStore, system),
			message.New(messageStore, messageService),
			payee.NewPayee(system, dapp, propertyStore, userStore, walletStore, marketStore, supplyStore, borrowStore, proposalStore, transactionStore, oracleSignerStore, proposalService, blockService, marketService, supplyService, borrowService, accountService, allowListService),
			syncer.New(walletStore, walletService, propertyStore),
			datadog.New(walletStore, propertyStore, messageService, provideDataDogConfig(cfg)),
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
