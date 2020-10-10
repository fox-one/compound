package cmd

import (
	"compound/worker"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "compound job worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		workers := []worker.Worker{}

		var g errgroup.Group
		for _, w := range workers {
			g.Go(func() error {
				return w.Run(ctx)
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
