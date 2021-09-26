package cmd

import (
	"compound/core"
	"compound/core/proposal"

	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/fox-one/pkg/qrcode"
	"github.com/spf13/cobra"
)

var allowListCmd = &cobra.Command{
	Use:     "allowlist",
	Aliases: []string{"al"},
	Short:   "allowlist cmd group",
	Example: heredoc.Doc(`
		$compound allowlist add --user {user_id} --scope {scope}
		$compound allowlist remove --user {user_id} --scope {scope}
	`),
}

var addAllowListCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"ad"},
	Short:   "add to allowlist",
	Long:    "enable user doing something of the specified scope, such as liquidation",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			panic(err)
		}

		scope, err := cmd.Flags().GetString("scope")
		if err != nil {
			panic(err)
		}

		if scope == "" {
			panic(errors.New("no scope specified"))
		}

		if !core.CheckScope(scope) {
			panic(errors.New("invalid scope"))
		}

		if userID == "" {
			//no user id, only scope
			scopeReq := proposal.ScopeReq{
				Scope: scope,
			}

			url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalAddScope, scopeReq)
			if err != nil {
				cmd.PrintErr(err)
				return
			}

			cmd.Println(url)
			qrcode.Fprint(cmd.OutOrStdout(), url)
		} else {
			allowListReq := proposal.AllowListReq{
				UserID: userID,
				Scope:  scope,
			}

			url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalAddAllowList, allowListReq)
			if err != nil {
				cmd.PrintErr(err)
				return
			}

			cmd.Println(url)
			qrcode.Fprint(cmd.OutOrStdout(), url)
		}
	},
}

var removeAllowListCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "remove from allowlist",
	Long:    "disable user doing something of the specified scope, such as liquidation",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			panic(err)
		}

		scope, err := cmd.Flags().GetString("scope")
		if err != nil {
			panic(err)
		}

		if scope == "" {
			panic(errors.New("no scope specified"))
		}

		if !core.CheckScope(scope) {
			panic(errors.New("invalid scope"))
		}

		if userID == "" {
			//no user id, only scope
			scopeReq := proposal.ScopeReq{
				Scope: scope,
			}

			url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalRemoveScope, scopeReq)
			if err != nil {
				cmd.PrintErr(err)
				return
			}

			cmd.Println(url)
			qrcode.Fprint(cmd.OutOrStdout(), url)
		} else {
			allowListReq := proposal.AllowListReq{
				UserID: userID,
				Scope:  scope,
			}

			url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalRemoveAllowList, allowListReq)
			if err != nil {
				cmd.PrintErr(err)
				return
			}

			cmd.Println(url)
			qrcode.Fprint(cmd.OutOrStdout(), url)
		}
	},
}

func init() {
	proposalCmd.AddCommand(allowListCmd)

	allowListCmd.AddCommand(addAllowListCmd)
	allowListCmd.AddCommand(removeAllowListCmd)

	addAllowListCmd.Flags().StringP("user", "u", "", "mixin user id")
	addAllowListCmd.Flags().StringP("scope", "s", "", "scope defined in compound: liquidation")

	removeAllowListCmd.Flags().StringP("user", "u", "", "mixin user id")
	removeAllowListCmd.Flags().StringP("scope", "s", "", "scope defined in compound: liquidation")
}
