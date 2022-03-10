package payee

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"strings"

	"github.com/fox-one/pkg/logger"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func (w *Payee) handleOutputV3(ctx context.Context, output *core.Output) (string, error) {
	log := logger.FromContext(ctx)
	if strings.ToLower(output.Memo) == "deposit" {
		return "", nil
	}

	var (
		action   core.ActionType
		followID string
	)
	message := w.decodeMemo(output.Memo)
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
		return followID, w.handleMakeProposal(ctx, output, message)
	case core.ActionTypeProposalShout:
		return followID, w.handleShoutProposal(ctx, output, message)
	case core.ActionTypeProposalVote:
		return followID, w.handleVoteProposal(ctx, output, message)
	default:
		user, err := w.userStore.Find(ctx, output.Sender)
		if err != nil {
			log.WithError(err).Errorln("users.Find")
			return "", err
		}

		if user.ID == 0 {
			//upsert user
			user = &core.User{
				UserID:    output.Sender,
				Address:   uuidutil.New(),
				AddressV0: core.BuildUserAddressV0(output.Sender),
			}
			if err = w.userStore.Create(ctx, user); err != nil {
				return "", err
			}
		}

		// handle user action
		return followID, w.handleUserAction(ctx, output, action, followID, message)
	}
}
