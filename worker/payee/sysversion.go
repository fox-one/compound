package payee

import (
	"compound/core"
	"compound/pkg/sysversion"
	"context"
	"fmt"

	"github.com/fox-one/pkg/logger"
)

func (w *Payee) loadSysVersion(ctx context.Context) error {
	log := logger.FromContext(ctx)

	sysversion, err := sysversion.ReadSysVersion(ctx, w.propertyStore)
	if err != nil {
		log.WithError(err).Errorln("sysversion.ReadSysVersion")
		return err
	}
	w.sysversion = sysversion
	return nil
}

func (w *Payee) validateNewSysVersion(ctx context.Context, ver int64) error {
	log := logger.FromContext(ctx)

	if ver <= w.sysversion {
		log.WithField("sysversion:new", ver).Infoln("skip")
		return errProposalSkip
	}

	if ver > core.SysVersion {
		err := fmt.Errorf("sys version: new version (%d) is greater than core.SysVersion (%d)", ver, core.SysVersion)
		log.WithError(err).Errorln("validateProposalAction fail")
		return err
	}
	return nil
}
