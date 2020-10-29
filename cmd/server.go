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

		mux := chi.NewMux()
		mux.Use(middleware.Recoverer)
		mux.Use(middleware.StripSlashes)
		mux.Use(cors.AllowAll().Handler)
		mux.Use(logger.WithRequestID)
		mux.Use(middleware.Logger)
		mux.Use(middleware.NewCompressor(5).Handler)

		{
			//hc
			mux.Mount("/hc", hc.Handle("1.0.0"))
		}

		{
			//rpc
			mux.Mount("/rpc/v1", nil)
		}

		{
			//restful api
			mux.Mount("/api/v1", rest.Handle(ctx))
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
	serverCmd.Flags().IntP("port", "p", 9000, "server port")
}
