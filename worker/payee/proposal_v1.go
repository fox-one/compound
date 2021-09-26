package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"compound/pkg/sysversion"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
)

var (
	errProposalSkip = fmt.Errorf("skip: invalid proposal")
)

func (w *Payee) handleMemberAction(ctx context.Context, output *core.Output, member string, body []byte) error {
	log := logger.FromContext(ctx)

	var action int
	body, err := mtg.Scan(body, &action)
	if err != nil {
		log.WithError(err).Debugln("scan proposal trace & action failed")
		return nil
	}

	log.WithField("trace", output.TraceID).Debugf("handle member action %s", core.ActionType(action).String())

	if core.ActionType(action) == core.ActionTypeProposalVote {
		var trace uuid.UUID
		if _, err := mtg.Scan(body, &trace); err != nil {
			return nil
		}
		return w.voteProposal(ctx, output, member, trace.String())
	}

	// new proposal
	p := &core.Proposal{
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.CreatedAt,
		TraceID:   output.TraceID,
		Creator:   output.Sender,
		AssetID:   output.AssetID,
		Amount:    output.Amount,
		Action:    core.ActionType(action),
		Version:   output.ID,
	}

	var content interface{}
	switch p.Action {
	case core.ActionTypeProposalAddMarket:
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
		log.Panicf("unknown proposal action %d", p.Action)
	}

	if _, err := mtg.Scan(body, content); err != nil {
		log.WithError(err).Debugln("decode proposal content failed")
	}

	if err := w.validateProposalAction(ctx, p.Action, content); err != nil {
		if err == errProposalSkip {
			return nil
		}
		return err
	}

	p.Content, _ = json.Marshal(content)

	if err := w.proposalStore.Create(ctx, p); err != nil {
		log.WithError(err).Errorln("proposals.Create")
		return err
	}

	if err := w.proposalService.ProposalCreated(ctx, p, member); err != nil {
		log.WithError(err).Errorln("notifier.ProposalVoted")
		return err
	}

	return w.loadSysVersion(ctx)
}

func (w *Payee) validateProposalAction(ctx context.Context, a core.ActionType, u interface{}) error {
	log := logger.FromContext(ctx).WithField("action", a.String())

	switch a {
	case core.ActionTypeProposalSetProperty:
		action := u.(*proposal.SetProperty)
		switch action.Key {
		case "":
			log.Infoln("skip: empty key")
			return errProposalSkip

		case sysversion.SysVersionKey:
			ver, err := strconv.ParseInt(action.Value, 10, 64)
			if err != nil {
				log.WithError(err).Infoln("skip")
				return errProposalSkip
			}

			return w.validateNewSysVersion(ctx, ver)
		}
	}
	return nil
}

func (w *Payee) voteProposal(ctx context.Context, output *core.Output, member string, traceID string) error {
	log := logger.FromContext(ctx).WithField("proposal", traceID)

	p, isNotFound, err := w.proposalStore.Find(ctx, traceID)
	if err != nil {
		// 如果 proposal 不存在，直接跳过
		if isNotFound {
			log.WithError(err).Debugln("proposal not found")
			return nil
		}

		log.WithError(err).Errorln("proposals.Find")
		return err
	}

	passed := p.PassedAt.Valid
	if passed && p.Version != output.ID {
		return nil
	}

	if !govalidator.IsIn(member, p.Votes...) {
		p.Votes = append(p.Votes, member)
		log.Infof("Proposal Voted by %s", member)

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
