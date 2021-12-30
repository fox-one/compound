package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"context"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/fox-one/pkg/logger"
	uuidutil "github.com/fox-one/pkg/uuid"
)

func (w *Payee) handleMakeProposalV1(ctx context.Context, output *core.Output, message []byte) error {
	log := logger.FromContext(ctx).WithField("handler", "proposal_make")

	var action core.ActionType
	{
		var v int
		body, err := mtg.Scan(message, &v)
		if err != nil {
			log.WithError(err).Errorln("scan action failed")
			return nil
		}
		action = core.ActionType(v)
		message = body
	}

	if !action.IsProposalAction() {
		return nil
	}

	proposal, err := w.buildProposalV1(ctx, output, action, message)
	if proposal == nil || err != nil {
		return err
	}

	if err := w.proposalStore.Create(ctx, proposal); err != nil {
		log.WithError(err).Errorln("proposals.Create")
		return err
	}

	if w.system.IsMember(output.Sender) {
		if err := w.proposalService.ProposalCreated(ctx, proposal, output.Sender, w.sysversion); err != nil {
			log.WithError(err).Errorln("proposalService.ProposalCreated")
			return err
		}
	} else if w.system.IsStaff(output.Sender) {
		if err := w.forwardProposalV1(ctx, output, proposal, core.ActionTypeProposalShout); err != nil {
			return err
		}
	}

	return nil
}

func (w *Payee) buildProposalV1(ctx context.Context, output *core.Output, action core.ActionType, message []byte) (*core.Proposal, error) {
	log := logger.FromContext(ctx)

	// new proposal
	p := &core.Proposal{
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.CreatedAt,
		TraceID:   output.TraceID,
		Creator:   output.Sender,
		AssetID:   output.AssetID,
		Amount:    output.Amount,
		Action:    action,
		Version:   output.ID,
	}

	var content encoding.BinaryUnmarshaler
	switch p.Action {
	case core.ActionTypeProposalUpsertMarket:
		content = &proposal.MarketReq{}
	case core.ActionTypeProposalWithdrawReserves:
		content = &proposal.WithdrawReq{}
	case core.ActionTypeProposalCloseMarket:
		content = &proposal.MarketStatusReq{}
	case core.ActionTypeProposalOpenMarket:
		content = &proposal.MarketStatusReq{}
	case core.ActionTypeProposalAddScope, core.ActionTypeProposalRemoveScope:
		content = &proposal.ScopeReq{}
	case core.ActionTypeProposalAddAllowList, core.ActionTypeProposalRemoveAllowList:
		content = &proposal.AllowListReq{}
	case core.ActionTypeProposalAddOracleSigner:
		content = &proposal.AddOracleSignerReq{}
	case core.ActionTypeProposalRemoveOracleSigner:
		content = &proposal.RemoveOracleSignerReq{}
	case core.ActionTypeProposalSetProperty:
		content = &proposal.SetProperty{}
	default:
		return nil, fmt.Errorf("unknown proposal action %d", p.Action)
	}

	if err := content.UnmarshalBinary(message); err != nil {
		log.WithError(err).Debugln("decode proposal content failed")
	}

	p.Content, _ = json.Marshal(content)
	if err := w.validateProposal(ctx, p); err != nil {
		if err == errProposalSkip {
			return nil, nil
		}
		return nil, err
	}

	return p, nil
}

func (w *Payee) forwardProposalV1(ctx context.Context, output *core.Output, p *core.Proposal, action core.ActionType) error {
	pid, _ := uuidutil.FromString(p.TraceID)
	data, _ := mtg.Encode(int(action), pid)
	memo := base64.StdEncoding.EncodeToString(data)

	if err := w.walletz.HandleTransfer(ctx, &core.Transfer{
		TraceID:   uuidutil.Modify(output.TraceID, p.TraceID+w.system.ClientID),
		AssetID:   w.system.VoteAsset,
		Amount:    w.system.VoteAmount,
		Threshold: w.system.Threshold,
		Opponents: w.system.MemberIDs,
		Memo:      memo,
	}); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.HandleTransfer")
		return err
	}

	return nil
}
