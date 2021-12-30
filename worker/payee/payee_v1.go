package payee

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
)

func (w *Payee) handleOutputV1(ctx context.Context, output *core.Output) error {
	log := logger.FromContext(ctx).
		WithField("output", output.TraceID).
		WithField("sysversion", w.sysversion)
	ctx = logger.WithContext(ctx, log)

	message := w.decodeMemo(output.Memo)

	// handle price provided by dirtoracle
	if priceData, err := w.decodePriceTransaction(ctx, message); err == nil {
		return w.handlePriceEvent(ctx, output, priceData)
	}

	if w.sysversion < 1 {
		// handle v0 member proposal action
		if member, action, body, err := core.DecodeMemberActionV0(message, w.system.Members); err == nil {
			return w.handleProposalActionV0(ctx, output, member, action, body)
		}
	}

	if output.Sender == "" {
		return nil
	}

	// 2. decode tx message
	if body, err := mtg.Decrypt(message, w.system.PrivateKey); err == nil {
		message = body
	}

	var action core.ActionType
	{
		var v int
		body, err := mtg.Scan(message, &v)
		if err != nil {
			log.WithError(err).Errorln("scan action failed")
			return nil
		}
		message = body
		action = core.ActionType(v)
	}

	log = log.WithField("action", action.String())
	ctx = logger.WithContext(ctx, log)

	switch action {
	case core.ActionTypeProposalMake:
		return w.handleMakeProposalV1(ctx, output, message)
	case core.ActionTypeProposalShout:
		return w.handleShoutProposal(ctx, output, message)
	case core.ActionTypeProposalVote:
		return w.handleVoteProposalV1(ctx, output, message)
	default:
		// transaction trace id as order id, different from output trace id
		var followID uuid.UUID
		message, err := mtg.Scan(message, &followID)
		if err != nil {
			log.WithError(err).Errorln("scan follow error")
			return nil
		}

		user, err := w.userStore.Find(ctx, output.Sender)
		if err != nil {
			log.WithError(err).Errorln("users.Find")
			return err
		}

		if user.ID == 0 {
			//upsert user
			user = &core.User{
				UserID:    output.Sender,
				Address:   uuidutil.New(),
				AddressV0: core.BuildUserAddressV0(output.Sender),
			}
			if err = w.userStore.Create(ctx, user); err != nil {
				return err
			}
		}

		// handle user action
		return w.handleUserAction(ctx, output, action, output.Sender, followID.String(), message)
	}
}
