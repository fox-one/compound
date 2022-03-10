package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"context"
	"database/sql"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/fox-one/pkg/logger"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
)

func (w *Payee) handleMakeProposal(ctx context.Context, output *core.Output, message []byte) error {
	log := logger.FromContext(ctx).WithField("handler", "proposal_make")

	var action core.ActionType
	message, err := mtg.Scan(message, &action)
	if err != nil {
		log.WithError(err).Errorln("scan action failed")
		return nil
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

	proposal, isNotFound, err := w.proposalStore.Find(ctx, trace.String())
	if err != nil {
		// 如果 proposal 不存在，直接跳过
		if isNotFound {
			log.WithError(err).Debugln("proposal not found")
			return nil
		}

		log.WithError(err).Errorln("proposals.Find")
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

func (w *Payee) buildProposal(ctx context.Context, output *core.Output, action core.ActionType, message []byte) (*core.Proposal, error) {
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
		return nil, err
	}

	return p, nil
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
