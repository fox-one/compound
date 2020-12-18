package cmd

import (
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		wg := sync.WaitGroup{}
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(n int) error {
				defer wg.Done()
				dur := time.Second
				for {
					select {
					case <-ctx.Done():
						cmd.Println(ctx.Err())
					case <-time.After(dur):
						cmd.Println(n, ":====>", time.Now())
					}
				}
			}(i)
		}

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
