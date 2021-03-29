package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"time"

	"github.com/fox-one/pkg/logger"
)

func (w *Payee) handleAddScopeEvent(ctx context.Context, p *core.Proposal, req proposal.ScopeReq, t time.Time) error {
	log := logger.FromContext(ctx).WithField("worker", "add-scope")
	log.Infoln("add operation scope:", req.Scope)

	return w.allowListService.AddAllowListScope(ctx, core.OperationScope(req.Scope))
}

func (w *Payee) handleRemoveScopeEvent(ctx context.Context, p *core.Proposal, req proposal.ScopeReq, t time.Time) error {
	log := logger.FromContext(ctx).WithField("worker", "remove-scope")
	log.Infoln("remove operation scope:", req.Scope)

	return w.allowListService.RemoveAllowListScope(ctx, core.OperationScope(req.Scope))
}

func (w *Payee) handleAddAllowListEvent(ctx context.Context, p *core.Proposal, req proposal.AllowListReq, t time.Time) error {
	log := logger.FromContext(ctx).WithField("worker", "add-allowlist")
	log.Infoln("add allow list:", req.Scope, ":", req.UserID)

	return w.allowListService.AddAllowList(ctx, req.UserID, core.OperationScope(req.Scope))
}

func (w *Payee) handleRemoveAllowListEvent(ctx context.Context, p *core.Proposal, req proposal.AllowListReq, t time.Time) error {
	log := logger.FromContext(ctx).WithField("worker", "remove-allowlist")
	log.Infoln("remove allow list:", req.Scope, ":", req.UserID)

	return w.allowListService.RemoveAllowList(ctx, req.UserID, core.OperationScope(req.Scope))
}
