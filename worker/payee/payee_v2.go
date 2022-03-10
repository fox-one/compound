package payee

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"

	"github.com/fox-one/pkg/logger"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handleOutputV2(ctx context.Context, output *core.Output) error {
	log := logger.FromContext(ctx)
	message := w.decodeMemo(output.Memo)

	var (
		action   core.ActionType
		followID string
	)
	if payload, err := core.DecodeTransactionAction(message); err == nil {
		if message, err = mtg.Scan(payload.Body, &action); err == nil {
			if follow, err := uuid.FromBytes(payload.FollowID); err == nil && follow != uuid.Nil {
				followID = follow.String()
			}
		}
	}

	log = log.WithFields(logrus.Fields{
		"action": action.String(),
		"follow": followID,
	})
	ctx = logger.WithContext(ctx, log)

	switch action {
	case core.ActionTypeProposalMake:
		return w.handleMakeProposal(ctx, output, message)
	case core.ActionTypeProposalShout:
		return w.handleShoutProposal(ctx, output, message)
	case core.ActionTypeProposalVote:
		return w.handleVoteProposal(ctx, output, message)
	default:
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
		return w.handleUserAction(ctx, output, action, followID, message)
	}
}
