package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
)

func (w *Payee) handleAddOracleSignerEvent(ctx context.Context, p *core.Proposal, req proposal.AddOracleSignerReq, output *core.Output) error {
	return w.oracleSignerStore.Save(ctx, req.UserID, req.PublicKey)
}

func (w *Payee) handleRemoveOracleSignerEvent(ctx context.Context, p *core.Proposal, req proposal.RemoveOracleSignerReq, output *core.Output) error {
	return w.oracleSignerStore.Delete(ctx, req.UserID)
}
