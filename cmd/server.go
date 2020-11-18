package cmd

import (
	"compound/handler/hc"
	"compound/handler/rest"
	"fmt"
	"net/http"

	"github.com/fox-one/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run compound api server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		config := provideConfig()
		db := provideDatabase()
		mainWallet := provideMainWallet()
		blockWallet := provideBlockWallet()

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

		mux := chi.NewMux()
		mux.Use(middleware.Recoverer)
		mux.Use(middleware.StripSlashes)
		mux.Use(cors.AllowAll().Handler)
		mux.Use(logger.WithRequestID)
		mux.Use(middleware.Logger)
		mux.Use(middleware.NewCompressor(5).Handler)

		{
			//hc
			mux.Mount("/hc", hc.Handle(rootCmd.Version))
		}

		{
			//rpc
			mux.Mount("/rpc/v1", nil)
		}

		{
			//restful api
			mux.Mount("/api/v1", rest.Handle(ctx, config, db, mainWallet, blockWallet, marketStore, supplyStore, borrowStore, walletService, blockService, priceService, accountService, marketService, supplyService, borrowService))
		}

		{
			//websocket
			mux.Mount("/ws/v1", nil)
		}

		port, _ := cmd.Flags().GetInt("port")
		addr := fmt.Sprintf(":%d", port)

		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		logrus.Infoln("serve at", addr)
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("server aborted")
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntP("port", "p", 80, "server port")
}
