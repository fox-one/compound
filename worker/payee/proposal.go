package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/mtg"
	"compound/pkg/sysversion"
	"context"
	"encoding/json"
	"strconv"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
)

func (w *Payee) handleShoutProposal(ctx context.Context, output *core.Output, message []byte) error {
	log := logger.FromContext(ctx).WithField("handler", "proposal_shout")

	if !w.system.IsMember(output.Sender) {
		return nil
	}

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

	if err := w.proposalService.ProposalCreated(ctx, proposal, output.Sender, w.sysversion); err != nil {
		log.WithError(err).Errorln("proposalService.ProposalCreated")
		return err
	}
	return nil
}

func (w *Payee) validateProposal(ctx context.Context, p *core.Proposal) error {
	log := logger.FromContext(ctx).WithField("action", p.Action.String())

	switch p.Action {
	case core.ActionTypeProposalSetProperty:
		var content proposal.SetProperty
		if err := json.Unmarshal([]byte(p.Content), &content); err != nil {
			log.WithError(err).Errorln("unmarshal SetProperty failed")
			return errProposalSkip
		}

		switch content.Key {
		case "":
			log.Infoln("skip: empty key")
			return errProposalSkip

		case sysversion.SysVersionKey:
			ver, err := strconv.ParseInt(content.Value, 10, 64)
			if err != nil {
				log.WithError(err).Infoln("skip")
				return errProposalSkip
			}

			return w.validateNewSysVersion(ctx, ver)
		}
	}
	return nil
}
