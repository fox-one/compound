package proposal

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"encoding/base64"
	"encoding/json"

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

func (p *service) ProposalCreated(ctx context.Context, proposal *core.Proposal, by string) error {
	buttons := generateButtons(ctx, p.marketStore, proposal)
	trace, err := uuid.FromString(proposal.TraceID)
	if err != nil {
		return err
	}
	uid, err := uuid.FromString(p.system.ClientID)
	if err != nil {
		return err
	}
	memo, err := mtg.Encode(uid, int(core.ActionTypeProposalVote), trace)
	if err != nil {
		return err
	}

	sign := mtg.Sign(memo, p.system.SignKey)
	memo = mtg.Pack(memo, sign)

	input := mixin.TransferInput{
		AssetID: p.system.VoteAsset,
		Amount:  p.system.VoteAmount,
		TraceID: uuid.Modify(proposal.TraceID, p.system.ClientID),
		Memo:    base64.StdEncoding.EncodeToString(memo),
	}
	input.OpponentMultisig.Receivers = p.system.MemberIDs()
	input.OpponentMultisig.Threshold = p.system.Threshold

	payment, err := p.client.VerifyPayment(ctx, input)
	if err != nil {
		return err
	}

	buttons = appendCode(buttons, "Vote", payment.CodeID)
	buttonsData, err := json.Marshal(buttons)
	if err != nil {
		return err
	}

	post := renderProposal(proposal)

	var messages []*core.Message
	for _, admin := range p.system.Admins {
		// post
		postMsg := &mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(p.system.ClientID, admin),
			MessageID:      uuid.Modify(proposal.TraceID, p.system.ClientID+admin),
			Category:       mixin.MessageCategoryPlainPost,
			Data:           base64.StdEncoding.EncodeToString(post),
		}

		// buttons
		buttonMsg := &mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(p.system.ClientID, admin),
			MessageID:      uuid.Modify(postMsg.MessageID, "buttons"),
			Category:       mixin.MessageCategoryAppButtonGroup,
			Data:           base64.StdEncoding.EncodeToString(buttonsData),
		}

		pmsg, err := core.BuildMessage(postMsg)
		if err != nil {
			return err
		}
		messages = append(messages, pmsg)

		bmsg, err := core.BuildMessage(buttonMsg)
		if err != nil {
			return err
		}
		messages = append(messages, bmsg)
	}

	return p.messages.Create(ctx, messages)
}

// ProposalApproved send proposal approved message to all the node managers
func (p *service) ProposalApproved(ctx context.Context, proposal *core.Proposal, by string) error {
	var messages []*core.Message

	post := renderApprovedBy(proposal, by)
	for _, admin := range p.system.Admins {
		quote := uuid.Modify(proposal.TraceID, p.system.ClientID+admin)
		msgReq := &mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(p.system.ClientID, admin),
			MessageID:      uuid.Modify(quote, "Approved By "+by),
			Category:       mixin.MessageCategoryPlainText,
			Data:           base64.StdEncoding.EncodeToString(post),
			QuoteMessageID: quote,
		}

		msg, err := core.BuildMessage(msgReq)
		if err != nil {
			return err
		}

		messages = append(messages, msg)
	}

	return p.messages.Create(ctx, messages)
}

// ProposalPassed send proposal approved message to all the node managers
func (p *service) ProposalPassed(ctx context.Context, proposal *core.Proposal) error {
	var messages []*core.Message

	post := []byte(passedTpl)
	for _, admin := range p.system.Admins {
		quote := uuid.Modify(proposal.TraceID, p.system.ClientID+admin)
		msgReq := &mixin.MessageRequest{
			RecipientID:    admin,
			ConversationID: mixin.UniqueConversationID(p.system.ClientID, admin),
			MessageID:      uuid.Modify(quote, "Proposal Passed"),
			Category:       mixin.MessageCategoryPlainText,
			Data:           base64.StdEncoding.EncodeToString(post),
			QuoteMessageID: quote,
		}

		msg, err := core.BuildMessage(msgReq)
		if err != nil {
			return err
		}

		messages = append(messages, msg)
	}

	return p.messages.Create(ctx, messages)
}
