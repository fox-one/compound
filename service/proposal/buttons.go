package parliament

import (
	"compound/core"

	"fmt"

	"github.com/fox-one/mixin-sdk-go"
)

func generateButtons(p *core.Proposal) mixin.AppButtonGroupMessage {
	var buttons mixin.AppButtonGroupMessage

	buttons = appendUser(buttons, "Creator", p.Creator)

	//TODO
	// switch p.Action {
	// case core.ProposalActionAddPair:
	// 	var action proposal.AddPair
	// 	_ = json.Unmarshal(p.Content, &action)
	// 	buttons = appendAsset(buttons, "Base", action.BaseAsset)
	// 	buttons = appendAsset(buttons, "Quote", action.QuoteAsset)
	// 	buttons = appendAsset(buttons, "LP", p.AssetID)
	// case core.ProposalActionWithdraw:
	// 	var action proposal.Withdraw
	// 	_ = json.Unmarshal(p.Content, &action)
	// 	buttons = appendAsset(buttons, "Asset", action.Asset)
	// 	buttons = appendUser(buttons, "Opponent", action.Opponent)
	// }

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
