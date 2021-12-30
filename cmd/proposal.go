/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"encoding"
	"encoding/base64"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/uuid"
	"github.com/spf13/cobra"
)

// proposalCmd represents the proposal command
var proposalCmd = &cobra.Command{
	Use:     "proposal <command>",
	Aliases: []string{"pp"},
	Short:   "Manager Proposals",
	Long:    "Work With rings Proposals",
}

func init() {
	rootCmd.AddCommand(proposalCmd)
}

func buildProposalTransferURL(ctx context.Context, system *core.System, dapp *mixin.Client, action core.ActionType, content encoding.BinaryMarshaler) (string, error) {
	data, err := mtg.Encode(core.ActionTypeProposalMake, action)
	if err != nil {
		return "", err
	}

	if content != nil {
		if bts, err := content.MarshalBinary(); err != nil {
			return "", err
		} else {
			data = append(data, bts...)
		}
	}

	data, err = core.TransactionAction{Body: data}.Encode()
	if err != nil {
		return "", err
	}

	input := mixin.TransferInput{
		AssetID: cfg.Group.Vote.Asset,
		Amount:  cfg.Group.Vote.Amount,

		TraceID: uuid.New(),
		Memo:    base64.StdEncoding.EncodeToString(data),
	}
	input.OpponentMultisig.Receivers = system.MemberIDs
	input.OpponentMultisig.Threshold = system.Threshold

	fmt.Println(input.Memo)
	payment, err := dapp.VerifyPayment(ctx, input)
	if err != nil {
		return "", err
	}

	return mixin.URL.Codes(payment.CodeID), nil
}
