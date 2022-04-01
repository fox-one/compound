package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"compound/pkg/mtg"
	"context"
	"database/sql"
	"encoding/base64"

	"github.com/asaskevich/govalidator"
	"github.com/fox-one/pkg/logger"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
)

func (w *Payee) handleMakeProposal(ctx context.Context, output *core.Output, message []byte) error {
	log := logger.FromContext(ctx).WithField("handler", "proposal_make")

	var action core.ActionType
	message, err := mtg.Scan(message, &action)
	if e := compound.Require(err == nil, "payee/mtgscan"); e != nil {
		log.WithError(err).Errorln("scan action failed")
		return e
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
		if err := w.forwardProposal(ctx, output, proposal, core.ActionTypeProposalShout); err != nil {
			return err
		}
	}

	return nil
}

func (w *Payee) handleVoteProposal(ctx context.Context, output *core.Output, message []byte) error {
	log := logger.FromContext(ctx).WithField("handler", "proposal_vote")

	var trace uuid.UUID
	if _, err := mtg.Scan(message, &trace); err != nil {
		log.WithError(err).Errorln("scan proposal trace failed")
		return nil
	}

	proposal, err := w.mustGetProposal(ctx, trace.String())
	if err != nil {
		return err
	}

	if w.system.IsStaff(output.Sender) {
		if err := w.forwardProposal(ctx, output, proposal, core.ActionTypeProposalVote); err != nil {
			return err
		}
		return nil
	} else if w.system.IsMember(output.Sender) {
		if err := w.validateProposal(ctx, proposal); err != nil {
			return err
		}

		if handled := proposal.PassedAt.Valid || govalidator.IsIn(output.Sender, proposal.Votes...); !handled {
			proposal.Votes = append(proposal.Votes, output.Sender)

			if err := w.proposalService.ProposalApproved(ctx, proposal, output.Sender, w.sysversion); err != nil {
				logger.FromContext(ctx).WithError(err).Errorln("proposalService.ProposalApproved")
				return err
			}

			if len(proposal.Votes) >= int(w.system.Threshold) {
				proposal.PassedAt = sql.NullTime{
					Time:  output.CreatedAt,
					Valid: true,
				}

				if err := w.proposalService.ProposalPassed(ctx, proposal, w.sysversion); err != nil {
					logger.FromContext(ctx).WithError(err).Errorln("proposalService.ProposalPassed")
					return err
				}
			}

			if err := w.proposalStore.Update(ctx, proposal, output.ID); err != nil {
				logger.FromContext(ctx).WithError(err).Errorln("proposals.Update")
				return err
			}
		}

		if proposal.PassedAt.Valid && proposal.Version == output.ID {
			return w.handlePassedProposal(ctx, proposal, output)
		}
	}
	return nil
}

func (w *Payee) forwardProposal(ctx context.Context, output *core.Output, p *core.Proposal, action core.ActionType) error {
	pid, _ := uuidutil.FromString(p.TraceID)
	data, _ := mtg.Encode(action, pid)
	data, _ = core.TransactionAction{Body: data}.Encode()
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
