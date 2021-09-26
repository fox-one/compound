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
	"compound/core/proposal"
	"compound/pkg/sysversion"
	"strconv"

	"github.com/fox-one/pkg/qrcode"
	"github.com/spf13/cobra"
)

// setpropertyCmd represents the setproperty command
var setpropertyCmd = &cobra.Command{
	Use:   "setproperty <key> <value>",
	Short: "create a proposal to set property",
	Args:  cobra.ExactValidArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		setPropReq := proposal.SetProperty{
			Key:   args[0],
			Value: args[1],
		}

		ctx := cmd.Context()
		system := provideSystem()
		dapp := provideDapp()

		if setPropReq.Key == "" {
			cmd.PrintErr("empty key")
			return
		} else if setPropReq.Key == sysversion.SysVersionKey {
			ver, err := strconv.ParseInt(setPropReq.Value, 10, 64)
			if err != nil {
				cmd.PrintErr(err)
				return
			}

			if ver > core.SysVersion {
				cmd.PrintErrf("sys version: new version (%d) is greater than core.SysVersion (%d)", ver, core.SysVersion)
				return
			}
		}

		url, err := buildProposalTransferURL(ctx, system, dapp.Client, core.ActionTypeProposalSetProperty, setPropReq)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		cmd.Println(url)
		qrcode.Fprint(cmd.OutOrStdout(), url)
	},
}

func init() {
	proposalCmd.AddCommand(setpropertyCmd)
}
