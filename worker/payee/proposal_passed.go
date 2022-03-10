package payee

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"encoding/json"
)

func (w *Payee) handlePassedProposal(ctx context.Context, p *core.Proposal, output *core.Output) error {
	if err := w.handlePassedProposalInternal(ctx, p, output); err != nil {
		return err
	}

	// TODO insert tx
	return nil
}

func (w *Payee) handlePassedProposalInternal(ctx context.Context, p *core.Proposal, output *core.Output) error {
	switch p.Action {
	case core.ActionTypeProposalUpsertMarket:
		var proposalReq proposal.MarketReq
		if err := json.Unmarshal(p.Content, &proposalReq); err != nil {
			return err
		}
		return w.handleMarketEvent(ctx, p, proposalReq, output)

	case core.ActionTypeProposalWithdrawReserves:
		var proposalReq proposal.WithdrawReq
		if err := json.Unmarshal(p.Content, &proposalReq); err != nil {
			return err
		}
		return w.handleWithdrawEvent(ctx, p, proposalReq, output)

	case core.ActionTypeProposalCloseMarket:
		var req proposal.MarketStatusReq
		if err := json.Unmarshal(p.Content, &req); err != nil {
			return err
		}
		return w.handleCloseMarketEvent(ctx, p, req, output)

	case core.ActionTypeProposalOpenMarket:
		var req proposal.MarketStatusReq
		if err := json.Unmarshal(p.Content, &req); err != nil {
			return err
		}
		return w.handleOpenMarketEvent(ctx, p, req, output)

	case core.ActionTypeProposalAddOracleSigner:
		var req proposal.AddOracleSignerReq
		if err := json.Unmarshal(p.Content, &req); err != nil {
			return err
		}
		return w.handleAddOracleSignerEvent(ctx, p, req, output)

	case core.ActionTypeProposalRemoveOracleSigner:
		var req proposal.RemoveOracleSignerReq
		if err := json.Unmarshal(p.Content, &req); err != nil {
			return err
		}
		return w.handleRemoveOracleSignerEvent(ctx, p, req, output)

	case core.ActionTypeProposalSetProperty:
		var req proposal.SetProperty
		if err := json.Unmarshal(p.Content, &req); err != nil {
			return err
		}
		return w.setProperty(ctx, output, p, req)
	}

	return nil
}
