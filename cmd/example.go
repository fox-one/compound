package cmd

import (
	"compound/core"
	"compound/pkg/mtg"
	"encoding/base64"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		log := logger.FromContext(ctx)

		system := provideSystem()

		memo := "CBm8+el8wqKoS+X0iwT5vXq8wXE94zqo/6pLXhQCCwcXd2OE1xiNYfXdSoMOjQv8OpK1L3GtPWJyOjVW1TS3++PDYvSkeAeXJcXJL1AHazZLqzrhkONXE9sLPmFJYQjw"
		bs, e := base64.StdEncoding.DecodeString(memo)
		if e != nil {
			panic(e)
		}

		actionType, body, err := core.DecodeUserTransactionAction(system.PrivateKey, bs)
		if err != nil {
			log.WithError(err).Errorln("DecodeTransactionAction error")
			panic(err)
		}

		cmd.Println("action:", actionType)
		var userID uuid.UUID
		// transaction trace id, different from output trace id
		var followID uuid.UUID
		body, err = mtg.Scan(body, &userID, &followID)
		if err != nil {
			log.WithError(err).Errorln("scan userID and followID error")
			panic(err)
		}

		cmd.Println("userID:", userID.String(), ":followID:", followID.String())
		// cmd.Println(ticker)
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
