package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"context"
	"database/sql"
	"encoding/json"

	"github.com/asaskevich/govalidator"
	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
)

func (w *Payee) handleProposalActionV0(ctx context.Context, output *core.Output, member *core.Member, action core.ActionType, body []byte) error {
	if action == core.ActionTypeProposalVote {
		return w.handleVoteProposalEventV0(ctx, output, member, body)
	}

	return w.handleCreateProposalEventV0(ctx, output, member, action, output.TraceID, body)
}

func (w *Payee) handleVoteProposalEventV0(ctx context.Context, output *core.Output, member *core.Member, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "proposal_vote")

	var traceID uuid.UUID
	_, err := mtg.Scan(body, &traceID)
	if err != nil {
		return nil
	}

	p, isRecordNotFound, err := w.proposalStore.Find(ctx, traceID.String())
	if err != nil {
		// 如果 proposal 不存在，直接跳过
		if isRecordNotFound {
			log.WithError(err).Debugln("proposal not found")
			return nil
		}

		log.WithError(err).Errorln("proposals.Find")
		return err
	}

	passed := p.PassedAt.Valid
	if passed && p.Version < output.ID {
		return nil
	}

	if !passed && !govalidator.IsIn(member.ClientID, p.Votes...) {
		p.Votes = append(p.Votes, member.ClientID)
		log.Infof("Proposal Voted by %s", member.ClientID)

		if err := w.proposalService.ProposalApproved(ctx, p, member.ClientID, w.sysversion); err != nil {
			log.WithError(err).Errorln("notifier.ProposalVoted")
			return err
		}

		if passed = len(p.Votes) >= int(w.system.Threshold); passed {
			p.PassedAt = sql.NullTime{
				Time:  output.CreatedAt,
				Valid: true,
			}

			log.Infof("Proposal Approved")
			if err := w.proposalService.ProposalPassed(ctx, p, w.sysversion); err != nil {
				log.WithError(err).Errorln("notifier.ProposalApproved")
				return err
			}
		}

		if err := w.proposalStore.Update(ctx, p, output.ID); err != nil {
			log.WithError(err).Errorln("proposals.Update")
			return err
		}
	}

	if passed {
		return w.handlePassedProposal(ctx, p, output)
	}

	return nil
}

func (w *Payee) handleCreateProposalEventV0(ctx context.Context, output *core.Output, member *core.Member, action core.ActionType, traceID string, body []byte) error {
	log := logger.FromContext(ctx).WithField("worker", "create_proposal")
	p := core.Proposal{
		TraceID:   traceID,
		Creator:   member.ClientID,
		AssetID:   output.AssetID,
		Amount:    output.Amount,
		Action:    action,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.CreatedAt,
	}

	switch p.Action {
	case core.ActionTypeProposalUpsertMarket:
		var content proposal.MarketReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal AddMarket content error")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalWithdrawReserves:
		var content proposal.WithdrawReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal WithdrawReserves content error")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalCloseMarket:
		var content proposal.MarketStatusReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal closeMarket content error")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalOpenMarket:
		var content proposal.MarketStatusReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal openMarket content error")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalAddScope, core.ActionTypeProposalRemoveScope:
		var content proposal.ScopeReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal scopereq content error")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalAddAllowList, core.ActionTypeProposalRemoveAllowList:
		var content proposal.AllowListReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal allowlist content error")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalAddOracleSigner:
		var content proposal.AddOracleSignerReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal add oracle signer content err")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalRemoveOracleSigner:
		var content proposal.RemoveOracleSignerReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal remove oracle signer content err")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	case core.ActionTypeProposalSetProperty:
		var content proposal.SetProperty
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal set property content err")
			return nil
		}
		bs, err := json.Marshal(content)
		if err != nil {
			return err
		}
		p.Content = bs
	default:
		log.Panicf("unknown proposal action %d", p.Action)
	}

	if err := w.proposalStore.Create(ctx, &p); err != nil {
		log.WithError(err).Errorln("proposal.create error")
		return err
	}

	if err := w.proposalService.ProposalCreated(ctx, &p, member.ClientID, w.sysversion); err != nil {
		log.WithError(err).Errorln("proposalCreated error")
		return err
	}

	return nil
}
