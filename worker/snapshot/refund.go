package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
	uuidutil "github.com/fox-one/pkg/uuid"
)

func (w *Payee) handleRefundEvent(ctx context.Context, output *core.Output, userID, followID string, errCode core.ErrorCode, msg string) error {
	log := logger.FromContext(ctx).WithField("worker", "refund")
	transferAction := core.TransferAction{
		Code:     int(errCode),
		Source:   core.ActionTypeRefundTransfer,
		FollowID: followID,
		Message:  msg,
	}
	memoStr, e := transferAction.Format()
	if e != nil {
		return e
	}

	log.Infof("userID:%s,followID:%s, error_code:%d", userID, followID, errCode)

	transfer := &core.Transfer{
		TraceID:   uuidutil.Modify(output.TraceID, "compound_refund"),
		Opponents: []string{userID},
		Threshold: 1,
		AssetID:   output.AssetID,
		Amount:    output.Amount,
		Memo:      memoStr,
	}

	if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{transfer}); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("walletStore.CreateTransfers")
		return err
	}

	return nil
}
