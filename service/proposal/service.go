package proposal

import (
	"compound/core"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/uuid"
)

// New new proposal service
func New(
	system *core.System,
	client *mixin.Client,
	marketStore core.IMarketStore,
	messages core.MessageStore,
) core.ProposalService {
	return &service{
		system:      system,
		client:      client,
		marketStore: marketStore,
		messages:    messages,
	}
}

type service struct {
	system      *core.System
	client      *mixin.Client
	marketStore core.IMarketStore
	messages    core.MessageStore
}

func (s *service) ProposalCreated(ctx context.Context, p *core.Proposal, by string, sysver int64) error {
	view := Proposal{
		Number: p.ID,
		Action: p.Action.String(),
		Info: []Item{
			{
				Key:   "action",
				Value: p.Action.String(),
			},
			{
				Key:   "id",
				Value: p.TraceID,
			},
			{
				Key:   "date",
				Value: p.CreatedAt.Format(time.RFC3339),
			},
			{
				Key:    "creator",
				Value:  s.fetchUserName(ctx, p.Creator),
				Action: userAction(p.Creator),
			},
			{
				Key:    "pay",
				Value:  fmt.Sprintf("%s %s", p.Amount, s.fetchAssetSymbol(ctx, p.AssetID)),
				Action: assetAction(p.AssetID),
			},
		},
	}

	view.Meta = s.renderProposalItems(ctx, p)

	items := append(view.Info, view.Meta...)
	voteAction, err := s.requestVoteAction(ctx, p, sysver)
	if err != nil {
		return err
	}

	items = append(items, Item{
		Key:    "Vote",
		Value:  "Vote",
		Action: voteAction,
	})

	buttons := generateButtons(items)
	buttonsData, _ := json.Marshal(buttons)
	post := execute("proposal_created", view)

	var messages []*core.Message
	for _, admin := range s.system.Admins {
		// post
		postMsg := &mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(s.system.ClientID, admin),
			MessageID:      uuid.Modify(p.TraceID, s.system.ClientID+admin),
			Category:       mixin.MessageCategoryPlainPost,
			Data:           base64.StdEncoding.EncodeToString(post),
		}

		// buttons
		buttonMsg := &mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(s.system.ClientID, admin),
			MessageID:      uuid.Modify(postMsg.MessageID, "buttons"),
			Category:       mixin.MessageCategoryAppButtonGroup,
			Data:           base64.StdEncoding.EncodeToString(buttonsData),
		}

		messages = append(messages, core.BuildMessage(postMsg), core.BuildMessage(buttonMsg))
	}

	return s.messages.Create(ctx, messages)
}

// ProposalApproved send proposal approved message to all the node managers
func (s *service) ProposalApproved(ctx context.Context, p *core.Proposal, by string, sysver int64) error {
	view := Proposal{
		ApprovedCount: len(p.Votes),
		ApprovedBy:    s.fetchUserName(ctx, by),
	}

	post := execute("proposal_approved", view)

	var messages []*core.Message
	for _, admin := range s.system.Admins {
		quote := uuid.Modify(p.TraceID, s.system.ClientID+admin)
		messages = append(messages, core.BuildMessage(&mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(s.system.ClientID, admin),
			MessageID:      uuid.Modify(quote, "Approved By "+by),
			Category:       mixin.MessageCategoryPlainText,
			Data:           base64.StdEncoding.EncodeToString(post),
			QuoteMessageID: quote,
		}))
	}

	return s.messages.Create(ctx, messages)
}

// ProposalPassed send proposal approved message to all the node managers
func (s *service) ProposalPassed(ctx context.Context, p *core.Proposal, sysver int64) error {
	post := execute("proposal_passed", nil)

	var messages []*core.Message
	for _, admin := range s.system.Admins {
		quote := uuid.Modify(p.TraceID, s.system.ClientID+admin)
		messages = append(messages, core.BuildMessage(&mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(s.system.ClientID, admin),
			MessageID:      uuid.Modify(quote, "Proposal Passed"),
			Category:       mixin.MessageCategoryPlainText,
			Data:           base64.StdEncoding.EncodeToString(post),
			QuoteMessageID: quote,
		}))
	}

	return s.messages.Create(ctx, messages)
}
