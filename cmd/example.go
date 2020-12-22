package cmd

import (
	"encoding/json"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:     "exam",
	Aliases: []string{"exam"},
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		// ctx := cmd.Context()
		// log := logger.FromContext(ctx)

		// system := provideSystem()

		// memo := "CBm8+el8wqKoS+X0iwT5vXq8wXE94zqo/6pLXhQCCwcXd2OE1xiNYfXdSoMOjQv8OpK1L3GtPWJyOjVW1TS3++PDYvSkeAeXJcXJL1AHazZLqzrhkONXE9sLPmFJYQjw"
		// bs, e := base64.StdEncoding.DecodeString(memo)
		// if e != nil {
		// 	panic(e)
		// }

		// actionType, body, err := core.DecodeUserTransactionAction(system.PrivateKey, bs)
		// if err != nil {
		// 	log.WithError(err).Errorln("DecodeTransactionAction error")
		// 	panic(err)
		// }

		// cmd.Println("action:", actionType)
		// var userID uuid.UUID
		// // transaction trace id, different from output trace id
		// var followID uuid.UUID
		// body, err = mtg.Scan(body, &userID, &followID)
		// if err != nil {
		// 	log.WithError(err).Errorln("scan userID and followID error")
		// 	panic(err)
		// }

		// cmd.Println("userID:", userID.String(), ":followID:", followID.String())
		// cmd.Println(ticker)

		m1 := `{"type":"multisig_utxo","user_id":"e15a8248-cd64-4e28-b10d-4907236e9fca","utxo_id":"1effd6f2-9b3d-3f19-9453-3e994c70eda4","asset_id":"4d8c508b-91c5-375b-92b0-ee702ed2dac5","transaction_hash":"044b89bc20ecdca968550ca0f0b296c42a21b3be1c81f04cedbe52ab02ebd0e3","output_index":0,"amount":"0.01","threshold":2,"members":["229fc7ac-9d09-4a6a-af5a-78f7439dce76","84a4db41-4992-4d35-aac7-987f965f0302","e15a8248-cd64-4e28-b10d-4907236e9fca"],"memo":"dxPkupEBwnGVRYEzZ6GvpqH/TUmjrsIB/5QcC7HoVEIriKCiWlBeSxrGM8QZAyZuDfl36gK6913/ddkW+0CcyFdSKPpMk928m6jGXlk5phU=","state":"unspent","created_at":"2020-12-22T02:48:32.565346Z","updated_at":"2020-12-22T02:48:32.565346Z","signed_by":"","signed_tx":""}`
		// m2 := `{"type":"multisig_utxo","user_id":"e15a8248-cd64-4e28-b10d-4907236e9fca","utxo_id":"07caeefa-52e3-3330-bc58-cc683fd330f0","asset_id":"965e5c6e-434c-3fa9-b780-c50f43cd955c","transaction_hash":"985d3be9f53b6c7e24bf63ded48b8df42311a5a2c004e772cd648dd2e9236c62","output_index":0,"amount":"0.00000128","threshold":2,"members":["229fc7ac-9d09-4a6a-af5a-78f7439dce76","84a4db41-4992-4d35-aac7-987f965f0302","e15a8248-cd64-4e28-b10d-4907236e9fca"],"memo":"merge for ec81384b-ffee-5e6e-8ec4-743100ed533c","state":"unspent","created_at":"2020-12-22T02:49:41.061108Z","updated_at":"2020-12-22T02:49:41.061108Z","signed_by":"","signed_tx":""}`

		var utxo mixin.MultisigUTXO
		if err := json.Unmarshal([]byte(m1), &utxo); err != nil {
			panic(err)
		}

		cmd.Println("utxo parse successful!")
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
