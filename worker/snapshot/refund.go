package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
)

// handle refund event
func (w *Payee) handleRefundEvent(ctx context.Context, output *core.Output, userID, followID string, origin core.ActionType, errCode core.ErrorCode) error {
	log := logger.FromContext(ctx).WithField("worker", "refund")

	transfer, e := core.NewRefundTransfer(output, userID, followID, origin, errCode)
	if e != nil {
		log.WithError(e).Errorln("new refund transfer error")
		return e
	}

	if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{transfer}); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("walletStore.CreateTransfers")
		return err
	}

	return nil
}
