package cmd

import (
	"compound/handler/hc"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/drone/signal"
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
			mux.Mount("/rpc", nil)
		}

		{
			//restful api
			mux.Mount("/api", nil)
		}

		{
			//websocket
			mux.Mount("/ws", nil)
		}

		port, _ := cmd.Flags().GetInt("port")
		addr := fmt.Sprintf(":%d", port)

		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		ctx, quit := context.WithCancel(ctx)
		done := make(chan struct{}, 1)
		signal.WithContextFunc(ctx, func() {
			quit()

			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				logrus.WithError(err).Error("graceful shutdown server failed")
			}

			close(done)
		})

		logrus.Infoln("serve at", addr)
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("server aborted")
		}

		<-done
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntP("port", "p", 9000, "server port")
}
