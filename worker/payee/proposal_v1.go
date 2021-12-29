package payee

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"encoding/base64"

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

	proposal, err := w.buildProposal(ctx, output, action, message)
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
