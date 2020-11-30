package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/logger"
)

var handleInjectMintTokenEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	log := logger.FromContext(ctx).WithField("worker", "mint")
	log.Infoln("inject mint token successful!")
	return nil
}
