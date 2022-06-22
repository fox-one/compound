package proposal

import (
	"compound/core"
	"compound/core/proposal"
	"context"
	"encoding/json"
	"fmt"
)

func (s *service) ListItems(ctx context.Context, p *core.Proposal) ([]core.ProposalItem, error) {
	var items []core.ProposalItem

	switch p.Action {
	case core.ActionTypeProposalUpsertMarket:
		var action proposal.MarketReq
		if err := json.Unmarshal(p.Content, &action); err != nil {
			return nil, err
		}

		items = []core.ProposalItem{
			{
				Key:   "symbol",
				Value: action.Symbol,
			},
			{
				Key:    "asset",
				Value:  action.AssetID,
				Hint:   s.fetchAssetSymbol(ctx, action.AssetID),
				Action: assetAction(action.AssetID),
			},
			{
				Key:    "ctoken",
				Value:  action.CTokenAssetID,
				Hint:   s.fetchAssetSymbol(ctx, action.CTokenAssetID),
				Action: assetAction(action.CTokenAssetID),
			},
			{
				Key:   "init_exchange",
				Value: action.InitExchange.String(),
			},
			{
				Key:   "reserve_factor",
				Value: action.ReserveFactor.String(),
			},
			{
				Key:   "liquidation_incentive",
				Value: action.LiquidationIncentive.String(),
			},
			{
				Key:   "collateral_factor",
				Value: action.CollateralFactor.String(),
			},
			{
				Key:   "base_rate",
				Value: action.BaseRate.String(),
			},
			{
				Key:   "borrow_cap",
				Value: action.BorrowCap.String(),
			},
			{
				Key:   "close_factor",
				Value: action.CloseFactor.String(),
			},
			{
				Key:   "multiplier",
				Value: action.Multiplier.String(),
			},
			{
				Key:   "jump_multiplier",
				Value: action.JumpMultiplier.String(),
			},
			{
				Key:   "kink",
				Value: action.Kink.String(),
			},
			{
				Key:   "price_threshold",
				Value: fmt.Sprint(action.PriceThreshold),
			},
			{
				Key:   "price",
				Value: action.Price.String(),
			},
			{
				Key:   "max_pledge",
				Value: action.MaxPledge.String(),
			},
		}
	case core.ActionTypeProposalWithdrawReserves:
		var action proposal.WithdrawReq
		if err := json.Unmarshal(p.Content, &action); err != nil {
			return nil, err
		}

		items = []core.ProposalItem{
			{
				Key:    "asset",
				Value:  action.Asset,
				Hint:   s.fetchAssetSymbol(ctx, action.Asset),
				Action: assetAction(action.Asset),
			},
			{
				Key:    "receiver",
				Value:  action.Opponent,
				Hint:   s.fetchUserName(ctx, action.Opponent),
				Action: userAction(action.Opponent),
			},
			{
				Key:   "amount",
				Value: action.Amount.String(),
			},
		}
	case core.ActionTypeProposalCloseMarket:
		var action proposal.MarketStatusReq
		if err := json.Unmarshal(p.Content, &action); err != nil {
			return nil, err
		}
		items = []core.ProposalItem{
			{
				Key:    "asset",
				Value:  action.AssetID,
				Hint:   s.fetchAssetSymbol(ctx, action.AssetID),
				Action: assetAction(action.AssetID),
			},
		}
	case core.ActionTypeProposalOpenMarket:
		var action proposal.MarketStatusReq
		if err := json.Unmarshal(p.Content, &action); err != nil {
			return nil, err
		}
		items = []core.ProposalItem{
			{
				Key:    "asset",
				Value:  action.AssetID,
				Hint:   s.fetchAssetSymbol(ctx, action.AssetID),
				Action: assetAction(action.AssetID),
			},
		}
	case core.ActionTypeProposalSetProperty:
		var action proposal.SetProperty
		if err := json.Unmarshal(p.Content, &action); err != nil {
			return nil, err
		}
		items = []core.ProposalItem{
			{
				Key:   action.Key,
				Value: action.Value,
			},
		}

	case core.ActionTypeProposalAddOracleSigner:
		var action proposal.AddOracleSignerReq
		if err := json.Unmarshal(p.Content, &action); err != nil {
			return nil, err
		}
		items = []core.ProposalItem{
			{
				Key:    "client",
				Value:  action.UserID,
				Hint:   s.fetchUserName(ctx, action.UserID),
				Action: userAction(action.UserID),
			},
			{
				Key:   "public_key",
				Value: action.PublicKey,
			},
		}
	}
	return items, nil
}

func (s *service) renderProposalItems(ctx context.Context, p *core.Proposal) []Item {
	items, _ := s.ListItems(ctx, p)

	results := make([]Item, len(items))
	for idx, item := range items {
		results[idx] = Item{
			Key:    item.Key,
			Value:  item.Value,
			Action: item.Action,
		}

		if item.Hint != "" {
			results[idx].Value = item.Hint
		}
	}

	return results
}
