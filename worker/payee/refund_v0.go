package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
)

// handle refund event
func (w *Payee) returnOrRefundError(ctx context.Context, err error, output *core.Output, userID, followID string, origin core.ActionType, errCode core.ErrorCode) error {
	var e compound.Error
	if errors.As(err, &e) && w.sysversion < 3 {
		return w.handleRefundEventV0(ctx, output, userID, followID, origin, errCode)
	}

	return err
}

// handle refund event
func (w *Payee) handleRefundEventV0(ctx context.Context, output *core.Output, userID, followID string, origin core.ActionType, errCode core.ErrorCode) error {
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
