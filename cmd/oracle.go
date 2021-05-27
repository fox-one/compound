package cmd

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/id"
	"compound/pkg/mtg"
	"encoding/base64"
	"fmt"

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

		clientID, err := uuid.FromString(system.ClientID)
		if err != nil {
			panic(err)
		}
		traceID, err := uuid.FromString(id.GenTraceID())
		if err != nil {
			panic(err)
		}

		user, err := cmd.Flags().GetString("user")
		if err != nil {
			panic(err)
		}
		publicKey, err := cmd.Flags().GetString("key")
		if err != nil {
			panic(err)
		}

		if user == "" || publicKey == "" {
			panic("no user or public key")
		}

		req := proposal.AddOracleSignerReq{
			UserID:    user,
			PublicKey: publicKey,
		}

		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalAddOracleSigner), req)
		if err != nil {
			panic(err)
		}

		sign := mtg.Sign(memo, system.SignKey)
		signedMemo := mtg.Pack(memo, sign)

		memoStr := base64.StdEncoding.EncodeToString(signedMemo)
		if len(memoStr) > 200 {
			memoStr = base64.StdEncoding.EncodeToString(memo)
		}

		fmt.Println("memo length:", len(memoStr))
		input := mixin.TransferInput{
			AssetID: system.VoteAsset,
			Amount:  system.VoteAmount,
			TraceID: traceID.String(),
			Memo:    memoStr,
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

		clientID, err := uuid.FromString(system.ClientID)
		if err != nil {
			panic(err)
		}
		traceID, err := uuid.FromString(id.GenTraceID())
		if err != nil {
			panic(err)
		}

		user, err := cmd.Flags().GetString("user")
		if err != nil {
			panic(err)
		}

		if user == "" {
			panic("no user")
		}

		req := proposal.RemoveOracleSignerReq{
			UserID: user,
		}

		memo, err := mtg.Encode(clientID, int(core.ActionTypeProposalRemoveOracleSigner), req)
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
