package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
)

var handleReserveEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	//保留金转账
	logger.FromContext(ctx).Infoln("reserve transfer")
	return nil
}
