package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/fox-one/pkg/logger"
)

func (w *Payee) handleVoteProposalEvent(ctx context.Context, output *core.Output, member *core.Member, traceID string) error {
	log := logger.FromContext(ctx).WithField("proposal", traceID)

	p, isRecordNotFound, err := w.proposalStore.Find(ctx, traceID)
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

	if !passed && !govalidator.IsIn(member.ClientID, p.Votes...) {
		p.Votes = append(p.Votes, member.ClientID)
		log.Infof("Proposal Voted by %s", member.ClientID)

		if err := w.proposalService.ProposalApproved(ctx, p, member); err != nil {
			log.WithError(err).Errorln("notifier.ProposalVoted")
			return err
		}

		if passed = len(p.Votes) >= int(w.system.Threshold); passed {
			p.PassedAt = sql.NullTime{
				Time:  output.CreatedAt,
				Valid: true,
			}

			log.Infof("Proposal Approved")
			if err := w.proposalService.ProposalPassed(ctx, p); err != nil {
				log.WithError(err).Errorln("notifier.ProposalApproved")
				return err
			}
		}

		if err := w.proposalStore.Update(ctx, p); err != nil {
			log.WithError(err).Errorln("proposals.Update")
			return err
		}
	}

	if passed {
		return w.handlePassedProposal(ctx, p, output.CreatedAt)
	}

	return nil
}

func (w *Payee) handleCreateProposalEvent(ctx context.Context, output *core.Output, member *core.Member, action core.ActionType, traceID string, body []byte) error {
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
	case core.ActionTypeProposalAddMarket:
		var content proposal.AddMarketReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal AddMarket content error")
			return nil
		}
		p.Content, _ = json.Marshal(content)
	case core.ActionTypeProposalUpdateMarket:
		var content proposal.UpdateMarketReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal UpdateMarket content error")
			return nil
		}
		p.Content, _ = json.Marshal(content)
	case core.ActionTypeProposalUpdateMarketAdvance:
		var content proposal.UpdateMarketAdvanceReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal UpdateMarketAdvance content error")
			return nil
		}
		p.Content, _ = json.Marshal(content)
	case core.ActionTypeProposalWithdrawReserves:
		var content proposal.WithdrawReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal WithdrawReserves content error")
			return nil
		}
		p.Content, _ = json.Marshal(content)
	case core.ActionTypeProposalCloseMarket:
		var content proposal.MarketStatusReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal closeMarket cotnent error")
			return nil
		}
		p.Content, _ = json.Marshal(content)
	case core.ActionTypeProposalOpenMarket:
		var content proposal.MarketStatusReq
		if _, err := mtg.Scan(body, &content); err != nil {
			log.WithError(err).Errorln("decode proposal openMarket cotnent error")
			return nil
		}
		p.Content, _ = json.Marshal(content)
	default:
		log.Warningln("invalid proposal:", p.Action)
		return nil
	}

	if err := w.proposalStore.Create(ctx, &p); err != nil {
		log.WithError(err).Errorln("proposal.create error")
		return err
	}

	if err := w.proposalService.ProposalCreated(ctx, &p, member); err != nil {
		log.WithError(err).Errorln("proposalCreated error")
		return err
	}

	return nil
}

func (w *Payee) handlePassedProposal(ctx context.Context, p *core.Proposal, t time.Time) error {
	switch p.Action {
	case core.ActionTypeProposalAddMarket:
		var proposalReq proposal.AddMarketReq
		_ = json.Unmarshal(p.Content, &proposalReq)
		return w.handleAddMarketEvent(ctx, p, proposalReq)

	case core.ActionTypeProposalUpdateMarket:
		var proposalReq proposal.UpdateMarketReq
		_ = json.Unmarshal(p.Content, &proposalReq)
		return w.handleUpdateMarketEvent(ctx, p, proposalReq, t)

	case core.ActionTypeProposalUpdateMarketAdvance:
		var proposalReq proposal.UpdateMarketAdvanceReq
		_ = json.Unmarshal(p.Content, &proposalReq)
		return w.handleUpdateMarketAdvanceEvent(ctx, p, proposalReq, t)

	case core.ActionTypeProposalWithdrawReserves:
		var proposalReq proposal.WithdrawReq
		_ = json.Unmarshal(p.Content, &proposalReq)
		return w.handleWithdrawEvent(ctx, p, proposalReq)
	case core.ActionTypeProposalCloseMarket:
		var req proposal.MarketStatusReq
		_ = json.Unmarshal(p.Content, &req)
		return w.handleCloseMarketEvent(ctx, p, req, t)
	case core.ActionTypeProposalOpenMarket:
		var req proposal.MarketStatusReq
		_ = json.Unmarshal(p.Content, &req)
		return w.handleOpenMarketEvent(ctx, p, req, t)
	}

	return nil
}
