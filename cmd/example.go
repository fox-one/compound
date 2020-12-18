package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		var g errgroup.Group
		for i := 0; i < 5; i++ {
			i := i
			g.Go(func() error {
				dur := time.Second
				for {
					select {
					case <-ctx.Done():
						cmd.Println(ctx.Err())
					case <-time.After(dur):
						cmd.Println(i, ":====>", time.Now())
					}
				}
			})
		}

		if err := g.Wait(); err != nil {
			cmd.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
