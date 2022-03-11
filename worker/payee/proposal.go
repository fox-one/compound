package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/compound"
	"compound/pkg/mtg"
	"compound/pkg/sysversion"
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/fox-one/pkg/logger"
	"github.com/gofrs/uuid"
	"github.com/pandodao/blst"
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
		{
			if err := compound.Require(json.Unmarshal([]byte(p.Content), &content) == nil, "payee/invalid-action"); err != nil {
				log.WithError(err).Errorln("unmarshal SetProperty failed")
				return err
			}
		}

		if err := compound.Require(content.Key != "", "payee/empty-key"); err != nil {
			log.WithError(err).Errorln("skip: empty key")
			return err
		}

		if content.Key == sysversion.SysVersionKey {
			ver, e := strconv.ParseInt(content.Value, 10, 64)
			if err := compound.Require(e == nil, "payee/invalid-sysversion"); err != nil {
				log.WithError(err).Errorln("validate sys version failed", ver)
				return err
			}

			return w.validateNewSysVersion(ctx, ver)
		}

	case core.ActionTypeProposalAddOracleSigner:
		var content proposal.AddOracleSignerReq
		{
			if err := compound.Require(json.Unmarshal([]byte(p.Content), &content) == nil, "payee/invalid-action"); err != nil {
				log.WithError(err).Errorln("unmarshal AddOracleSignerReq failed")
				return err
			}
		}

		bts, err := base64.StdEncoding.DecodeString(content.PublicKey)
		if e := compound.Require(
			err == nil,
			"payee/invalid-oracle-signer",
		); e != nil {
			return e
		}

		pub := blst.PublicKey{}
		if err := compound.Require(
			pub.FromBytes(bts) == nil,
			"payee/invalid-oracle-signer",
		); err != nil {
			return err
		}
	}
	return nil
}
