package snapshot

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/sysversion"
	"context"
	"strconv"

	"github.com/fox-one/pkg/logger"
)

func (w *Payee) setProperty(ctx context.Context, output *core.Output, _ *core.Proposal, action proposal.SetProperty) error {
	log := logger.FromContext(ctx)

	if action.Key == "" {
		return nil
	}

	if action.Key == sysversion.SysVersionKey {
		ver, err := strconv.ParseInt(action.Value, 10, 64)
		if err != nil {
			log.WithError(err).Errorln("skip: parse sysversion failled", action.Value)
			return nil
		}

		if err := w.validateNewSysVersion(ctx, ver); err != nil {
			if err == errProposalSkip {
				return nil
			}
			return err
		}

		if err := w.migrateSystem(ctx, ver, output.ID); err != nil {
			return err
		}
	}

	if err := w.propertyStore.Save(ctx, action.Key, action.Value); err != nil {
		log.WithError(err).Errorln("update properties", action.Key, action.Value)
		return err
	}

	return nil
}

func (w *Payee) migrateSystem(ctx context.Context, sysversion, version int64) error {
	if w.sysversion < 1 {
		if err := w.migrateV1(ctx, version); err != nil {
			return err
		}
	}

	return nil
}

func (w *Payee) migrateV1(ctx context.Context, version int64) error {
	return nil
}
