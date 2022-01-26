package proposal

import (
	"compound/core"
	"compound/pkg/mtg"
	"context"
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/uuid"
)

func (s *service) fetchAssetSymbol(ctx context.Context, assetID string) string {
	if uuid.IsNil(assetID) {
		return "ALL"
	}

	asset, err := s.client.ReadAsset(ctx, assetID)
	if err != nil {
		return assetID
	}

	return asset.Symbol
}

func (s *service) fetchUserName(ctx context.Context, userID string) string {
	user, err := s.client.ReadUser(ctx, userID)
	if err != nil {
		return userID
	}

	return user.FullName
}

func (s *service) requestVoteAction(ctx context.Context, proposal *core.Proposal, sysver int64) (string, error) {
	trace, _ := uuid.FromString(proposal.TraceID)

	var memo []byte
	var err error
	if sysver < 2 {
		memo, err = mtg.Encode(int(core.ActionTypeProposalVote), trace)
		if err != nil {
			return "", err
		}
	} else {
		memo, err = mtg.Encode(core.ActionTypeProposalVote, trace)
		if err != nil {
			return "", err
		}
		memo, err = core.TransactionAction{Body: memo}.Encode()
		if err != nil {
			return "", err
		}
	}

	input := mixin.TransferInput{
		AssetID: s.system.VoteAsset,
		Amount:  s.system.VoteAmount,
		TraceID: uuid.Modify(proposal.TraceID, s.system.ClientID),
		Memo:    base64.StdEncoding.EncodeToString(memo),
	}
	input.OpponentMultisig.Receivers = s.system.MemberIDs
	input.OpponentMultisig.Threshold = s.system.Threshold

	payment, err := s.client.VerifyPayment(ctx, input)
	if err != nil {
		return "", err
	}
	return paymentAction(payment.CodeID), nil
}
