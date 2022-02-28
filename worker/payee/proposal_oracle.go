package payee

import (
	"compound/core"
	"compound/core/proposal"
	"context"

	"github.com/fox-one/pkg/logger"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handleAddOracleSignerEvent(ctx context.Context, p *core.Proposal, req proposal.AddOracleSignerReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "add-oracle-signer",
		"signer":   req.UserID,
	})
	if err := w.oracleSignerStore.Save(ctx, req.UserID, req.PublicKey); err != nil {
		log.WithError(err).Errorln("add oracle signer failed")
		return err
	}
	return nil
}

func (w *Payee) handleRemoveOracleSignerEvent(ctx context.Context, p *core.Proposal, req proposal.RemoveOracleSignerReq, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "remove-oracle-signer",
		"signer":   req.UserID,
	})
	if err := w.oracleSignerStore.Delete(ctx, req.UserID); err != nil {
		log.WithError(err).Errorln("remove oracle signer failed")
		return err
	}
	return nil
}
