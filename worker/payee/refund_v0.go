package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"

	"github.com/fox-one/pkg/logger"
)

// handle refund event
func (w *Payee) handleRefundError(ctx context.Context, err error, output *core.Output, userID, followID string, origin core.ActionType, errCode core.ErrorCode) error {
	if _, ok := err.(compound.Error); ok && w.sysversion < 3 {
		return w.handleRefundEventV0(ctx, output, userID, followID, origin, errCode)
	}

	return err
}

// handle refund event
func (w *Payee) handleRefundEventV0(ctx context.Context, output *core.Output, userID, followID string, origin core.ActionType, errCode core.ErrorCode) error {
	log := logger.FromContext(ctx).WithField("worker", "refund")

	if w.sysversion < 3 {
		return nil
	}

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
