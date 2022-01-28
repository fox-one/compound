package views

import (
	"compound/core"
	"time"

	"github.com/shopspring/decimal"
)

type (
	ProposalItem struct {
		// Key is the parameter name
		Key string `json:"key,omitempty"`
		// Value the proposal applied
		Value string `json:"value,omitempty"`
		// Hint the parameter hint
		Hint string `json:"hint,omitempty"`
		// Action the value applied
		Action string `json:"action,omitempty"`
	}

	Proposal struct {
		ID        string          `json:"id,omitempty"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
		PassedAt  *time.Time      `json:"passed_at,omitempty"`
		Creator   string          `json:"creator,omitempty"`
		AssetID   string          `json:"asset_id,omitempty"`
		Amount    decimal.Decimal `json:"amount,omitempty"`
		Action    string          `json:"action,omitempty"`
		Votes     []string        `json:"votes,omitempty"`
		Items     []ProposalItem  `json:"items,omitempty"`
	}
)

func ProposalItemView(p core.ProposalItem) ProposalItem {
	return ProposalItem{
		Key:    p.Key,
		Value:  p.Value,
		Hint:   p.Hint,
		Action: p.Action,
	}
}

func ProposalItemViews(pitems []core.ProposalItem) []ProposalItem {
	var items = make([]ProposalItem, len(pitems))
	for i, item := range pitems {
		items[i] = ProposalItemView(item)
	}
	return items
}

func ProposalView(p core.Proposal) Proposal {
	view := Proposal{
		ID:        p.TraceID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Creator:   p.Creator,
		AssetID:   p.AssetID,
		Amount:    p.Amount,
		Action:    p.Action.String(),
		Votes:     p.Votes,
	}
	if p.PassedAt.Valid {
		view.PassedAt = &p.PassedAt.Time
	}
	return view
}

func ProposalViews(ps []*core.Proposal) []Proposal {
	var items = make([]Proposal, len(ps))
	for i, item := range ps {
		items[i] = ProposalView(*item)
	}
	return items
}
