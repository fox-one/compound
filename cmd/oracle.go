package cmd

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/id"
	"compound/pkg/mtg"
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
)

var addOracleSignerCmd = &cobra.Command{
	Use:     "add-oracle-signer",
	Aliases: []string{"aos"},
	Short:   "add oracle signer",
	Long: `flags->
	user: oracle signer user id
	key: public key of signer`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		user, _ := cmd.Flags().GetString("user")
		publicKey, _ := cmd.Flags().GetString("key")

		if user == "" || publicKey == "" {
			panic("no user or public key")
		}

		req := proposal.AddOracleSignerReq{
			UserID:    user,
			PublicKey: publicKey,
		}

		memo, err := mtg.Encode(clientID, traceID, int(core.ActionTypeProposalAddOracleSigner), req)
		if err != nil {
			panic(err)
		}

		sign := mtg.Sign(memo, system.SignKey)
		memo = mtg.Pack(memo, sign)

		input := mixin.TransferInput{
			AssetID: system.VoteAsset,
			Amount:  system.VoteAmount,
			TraceID: traceID.String(),
			Memo:    base64.StdEncoding.EncodeToString(memo),
		}
		input.OpponentMultisig.Receivers = system.MemberIDs()
		input.OpponentMultisig.Threshold = system.Threshold

		payment, err := dapp.Client.VerifyPayment(ctx, input)
		if err != nil {
			panic(err)
		}

		url := mixin.URL.Codes(payment.CodeID)
		cmd.Println(url)

		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

var removeOracleSignerCmd = &cobra.Command{
	Use:     "rm-oracle-signer",
	Aliases: []string{"ros"},
	Short:   "remove oracle signer",
	Long: `flags->
	user: oracle signer user id`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		clientID, _ := uuid.FromString(system.ClientID)
		traceID, _ := uuid.FromString(id.GenTraceID())

		user, _ := cmd.Flags().GetString("user")

		if user == "" {
			panic("no user")
		}

		req := proposal.RemoveOracleSignerReq{
			UserID: user,
		}

		memo, err := mtg.Encode(clientID, traceID, int(core.ActionTypeProposalRemoveOracleSigner), req)
		if err != nil {
			panic(err)
		}

		sign := mtg.Sign(memo, system.SignKey)
		memo = mtg.Pack(memo, sign)

		input := mixin.TransferInput{
			AssetID: system.VoteAsset,
			Amount:  system.VoteAmount,
			TraceID: traceID.String(),
			Memo:    base64.StdEncoding.EncodeToString(memo),
		}
		input.OpponentMultisig.Receivers = system.MemberIDs()
		input.OpponentMultisig.Threshold = system.Threshold

		payment, err := dapp.Client.VerifyPayment(ctx, input)
		if err != nil {
			panic(err)
		}

		url := mixin.URL.Codes(payment.CodeID)
		cmd.Println(url)

		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

func init() {
	rootCmd.AddCommand(addOracleSignerCmd)
	rootCmd.AddCommand(removeOracleSignerCmd)

	addOracleSignerCmd.Flags().String("user", "", "oracle signer user id")
	addOracleSignerCmd.Flags().String("key", "", "publick key of signer")

	removeOracleSignerCmd.Flags().String("user", "", "oracle signer user id")
}
