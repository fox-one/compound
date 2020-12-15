package parliament

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"encoding/json"
	"strings"

	"fmt"

	"github.com/fox-one/mixin-sdk-go"
)

func generateButtons(ctx context.Context, marketStore core.IMarketStore, p *core.Proposal) mixin.AppButtonGroupMessage {
	var buttons mixin.AppButtonGroupMessage

	buttons = appendUser(buttons, "Creator", p.Creator)

	switch p.Action {
	case core.ActionTypeProposalAddMarket:
		var action proposal.AddMarketReq
		_ = json.Unmarshal(p.Content, &action)
		buttons = appendAsset(buttons, "Asset", action.AssetID)
		buttons = appendAsset(buttons, "CToken", action.CTokenAssetID)
	case core.ActionTypeProposalUpdateMarket:
		var action proposal.UpdateMarketReq
		_ = json.Unmarshal(p.Content, &action)
		symbol := strings.ToUpper(action.Symbol)
		market, _, e := marketStore.FindBySymbol(ctx, symbol)
		if e != nil {
			return buttons
		}
		buttons = appendAsset(buttons, "Asset", market.AssetID)
	case core.ActionTypeProposalWithdrawReserves:
		var action proposal.WithdrawReq
		_ = json.Unmarshal(p.Content, &action)
		buttons = appendAsset(buttons, "Asset", action.Asset)
		buttons = appendUser(buttons, "Opponent", action.Opponent)
	}

	return buttons
}

func appendAsset(buttons mixin.AppButtonGroupMessage, label, id string) mixin.AppButtonGroupMessage {
	return append(buttons, mixin.AppButtonMessage{
		Label:  label,
		Action: fmt.Sprintf("https://mixin.one/snapshots/%s", id),
		Color:  randomHexColor(),
	})
}

func appendUser(buttons mixin.AppButtonGroupMessage, label, id string) mixin.AppButtonGroupMessage {
	return append(buttons, mixin.AppButtonMessage{
		Label:  label,
		Action: fmt.Sprintf("mixin://users/%s", id),
		Color:  randomHexColor(),
	})
}

func appendCode(buttons mixin.AppButtonGroupMessage, label, id string) mixin.AppButtonGroupMessage {
	return append(buttons, mixin.AppButtonMessage{
		Label:  label,
		Action: fmt.Sprintf("mixin://codes/%s", id),
		Color:  randomHexColor(),
	})
}
