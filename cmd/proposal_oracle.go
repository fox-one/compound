package cmd

import (
	"compound/core"
	"compound/core/proposal"

	"github.com/fox-one/pkg/qrcode"
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

		url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalAddOracleSigner, req)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

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

		url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalRemoveOracleSigner, req)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		cmd.Println(url)
		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

func init() {
	proposalCmd.AddCommand(addOracleSignerCmd)
	proposalCmd.AddCommand(removeOracleSignerCmd)

	addOracleSignerCmd.Flags().String("user", "", "oracle signer user id")
	addOracleSignerCmd.Flags().String("key", "", "publick key of signer")

	removeOracleSignerCmd.Flags().String("user", "", "oracle signer user id")
}
