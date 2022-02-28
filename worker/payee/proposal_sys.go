package payee

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/sysversion"
	"context"
	"fmt"
	"strconv"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/sirupsen/logrus"
)

var (
	errProposalSkip = fmt.Errorf("skip: invalid proposal")
)

func (w *Payee) setProperty(ctx context.Context, output *core.Output, _ *core.Proposal, action proposal.SetProperty) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"proposal": "set-property",
		"key":      action.Key,
	})
	ctx = logger.WithContext(ctx, log)

	if action.Key == "" {
		log.Errorln("skip: empty key")
		return errProposalSkip
	}

	if action.Key == sysversion.SysVersionKey {
		ver, err := strconv.ParseInt(action.Value, 10, 64)
		if err != nil {
			log.WithError(err).Errorln("skip: parse sysversion failled", action.Value)
			return errProposalSkip
		}

		if err := w.validateNewSysVersion(ctx, ver); err != nil {
			if err == errProposalSkip {
				log.Errorln("skip: invalid value", action.Value)
				return errProposalSkip
			}
			log.WithError(err).Errorln("validate sys version failed", ver)
			return err
		}

		if err := w.migrateSystem(ctx, ver, output.ID); err != nil {
			log.WithError(err).WithField("new-version", ver).Errorln("migrate system")
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
	log := logger.FromContext(ctx)

	var from uint64
	const limit = 500
	for {
		users, err := w.userStore.List(ctx, from, limit)
		if err != nil {
			log.WithError(err).Errorln("users.List")
			return err
		}

		var updates = make([]*core.User, 0, len(users))
		for _, user := range users {
			from = user.ID
			addressV0 := core.BuildUserAddressV0(user.UserID)
			if user.Address == addressV0 {
				user.AddressV0 = addressV0
				user.Address = uuid.New()
				updates = append(updates, user)
			}
		}

		if err := w.userStore.MigrateToV1(ctx, updates); err != nil {
			log.WithError(err).Errorln("users.MigrateToV1")
			return err
		}

		if len(users) < limit {
			break
		}
	}
	return nil
}
