package payee

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"encoding/json"
)

func (w *Payee) handlePassedProposal(ctx context.Context, p *core.Proposal, output *core.Output) error {
	switch p.Action {
	case core.ActionTypeProposalAddMarket:
		var proposalReq proposal.MarketReq
		err := json.Unmarshal(p.Content, &proposalReq)
		if err != nil {
			return err
		}
		return w.handleMarketEvent(ctx, p, proposalReq, output)

	case core.ActionTypeProposalWithdrawReserves:
		var proposalReq proposal.WithdrawReq
		err := json.Unmarshal(p.Content, &proposalReq)
		if err != nil {
			return err
		}
		return w.handleWithdrawEvent(ctx, p, proposalReq, output)

	case core.ActionTypeProposalCloseMarket:
		var req proposal.MarketStatusReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleCloseMarketEvent(ctx, p, req, output)

	case core.ActionTypeProposalOpenMarket:
		var req proposal.MarketStatusReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleOpenMarketEvent(ctx, p, req, output)

	case core.ActionTypeProposalAddScope:
		var req proposal.ScopeReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleAddScopeEvent(ctx, p, req, output)

	case core.ActionTypeProposalRemoveScope:
		var req proposal.ScopeReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleRemoveScopeEvent(ctx, p, req, output)

	case core.ActionTypeProposalAddAllowList:
		var req proposal.AllowListReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleAddAllowListEvent(ctx, p, req, output)

	case core.ActionTypeProposalRemoveAllowList:
		var req proposal.AllowListReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleRemoveAllowListEvent(ctx, p, req, output)

	case core.ActionTypeProposalAddOracleSigner:
		var req proposal.AddOracleSignerReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleAddOracleSignerEvent(ctx, p, req, output)

	case core.ActionTypeProposalRemoveOracleSigner:
		var req proposal.RemoveOracleSignerReq
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.handleRemoveOracleSignerEvent(ctx, p, req, output)

	case core.ActionTypeProposalSetProperty:
		var req proposal.SetProperty
		err := json.Unmarshal(p.Content, &req)
		if err != nil {
			return err
		}
		return w.setProperty(ctx, output, p, req)
	}

	return nil
}
