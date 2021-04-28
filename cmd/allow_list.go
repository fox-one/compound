package cmd

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/id"
	"compound/pkg/mtg"
	"context"
	"encoding/base64"

	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/qrcode"
	"github.com/gofrs/uuid"
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
		buildProposalTransfer(cmd, func(ctx context.Context, clientID uuid.UUID) ([]byte, error) {
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

			var memo []byte
			if userID == "" {
				//no user id, only scope
				scopeReq := proposal.ScopeReq{
					Scope: scope,
				}

				memo, err = mtg.Encode(clientID, int(core.ActionTypeProposalAddScope), scopeReq)
			} else {
				allowListReq := proposal.AllowListReq{
					UserID: userID,
					Scope:  scope,
				}

				memo, err = mtg.Encode(clientID, int(core.ActionTypeProposalAddAllowList), allowListReq)
			}

			return memo, err
		})
	},
}

var removeAllowListCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "remove from allowlist",
	Long:    "disable user doing something of the specified scope, such as liquidation",
	Run: func(cmd *cobra.Command, args []string) {
		buildProposalTransfer(cmd, func(ctx context.Context, clientID uuid.UUID) ([]byte, error) {
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

			var memo []byte
			if userID == "" {
				//no user id, only scope
				scopeReq := proposal.ScopeReq{
					Scope: scope,
				}

				memo, err = mtg.Encode(clientID, int(core.ActionTypeProposalRemoveScope), scopeReq)
			} else {
				allowListReq := proposal.AllowListReq{
					UserID: userID,
					Scope:  scope,
				}

				memo, err = mtg.Encode(clientID, int(core.ActionTypeProposalRemoveAllowList), allowListReq)
			}

			return memo, err
		})
	},
}

// BuildMemoFunc build memo func
type BuildMemoFunc func(ctx context.Context, clientID uuid.UUID) ([]byte, error)

func buildProposalTransfer(cmd *cobra.Command, f BuildMemoFunc) {
	ctx := cmd.Context()
	system := provideSystem()
	dapp := provideDapp()

	clientID, _ := uuid.FromString(system.ClientID)
	traceID, _ := uuid.FromString(id.GenTraceID())

	memo, err := f(ctx, clientID)
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
}

func init() {
	rootCmd.AddCommand(allowListCmd)

	allowListCmd.AddCommand(addAllowListCmd)
	allowListCmd.AddCommand(removeAllowListCmd)

	addAllowListCmd.Flags().StringP("user", "u", "", "mixin user id")
	addAllowListCmd.Flags().StringP("scope", "s", "", "scope defined in compound: liquidation")

	removeAllowListCmd.Flags().StringP("user", "u", "", "mixin user id")
	removeAllowListCmd.Flags().StringP("scope", "s", "", "scope defined in compound: liquidation")
}
