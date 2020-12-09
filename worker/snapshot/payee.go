package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"compound/worker"
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
	"github.com/robfig/cron/v3"
)

const (
	checkpointKey = "outputs_checkpoint"
	limit         = 500
)

// Payee payee worker
type Payee struct {
	worker.BaseJob
	system      *core.System
	property    property.Store
	walletStore core.WalletStore
}

// NewPayee new payee
func NewPayee(location string,
	system *core.System,
	property property.Store) *Payee {
	payee := Payee{
		property: property,
	}

	l, _ := time.LoadLocation(location)
	payee.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 100ms"
	payee.Cron.AddFunc(spec, payee.Run)
	payee.OnWork = func() error {
		return payee.onWork(context.Background())
	}

	return &payee
}

func (w *Payee) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "payee")

	v, err := w.property.Get(ctx, checkpointKey)
	if err != nil {
		log.WithError(err).Errorln("property.Get error")
		return err
	}

	outputs, err := w.walletStore.List(ctx, v.Int64(), limit)
	if err != nil {
		log.WithError(err).Errorln("walletStore.List")
		return err
	}

	if len(outputs) <= 0 {
		return errors.New("no more outputs")
	}

	for _, u := range outputs {
		if err := w.handleOutput(ctx, u); err != nil {
			return err
		}

		if err := w.property.Save(ctx, checkpointKey, u.ID); err != nil {
			log.WithError(err).Errorln("property.Save:", u.ID)
			return err
		}
	}

	return nil
}

func (w *Payee) handleOutput(ctx context.Context, output *core.Output) error {
	log := logger.FromContext(ctx).WithField("output", output.TraceID)
	ctx = logger.WithContext(ctx, log)

	message := w.decodeMemo(output.Memo)

	// handle member vote action
	if member, body, err := core.DecodeMemberProposalTransactionAction(message, w.system.Members); err == nil {
		return w.handleMemberAction(ctx, output, member, body)
	}

	// handle user action
	actionType, body, err := core.DecodeUserTransactionAction(w.system.PrivateKey, message)
	if err != nil {
		log.WithError(err).Errorln("DecodeTransactionAction error")
		return nil
	}

	var userID uuid.UUID
	// transaction trace id, different from output trace id
	var followID uuid.UUID
	body, err = mtg.Scan(body, &userID, &followID)
	if err != nil {
		log.WithError(err).Errorln("scan userID and followID error")
		return nil
	}

	return w.handleUserAction(ctx, output, actionType, userID, followID, body)
}

func (w *Payee) handleMemberAction(ctx context.Context, output *core.Output, member *core.Member, body []byte) error {
	return nil
}

func (w *Payee) handleUserAction(ctx context.Context, output *core.Output, actionType core.ActionType, userID, followID uuid.UUID, body []byte) error {
	return nil
}



func (w *Payee) transferOut(ctx context.Context) error {
	return nil
}

func (w *Payee) refundOutput(ctx context.Context, output *core.Output, userID, followID, msg string) error {
	// TODO memo should be formated
	transfer := &core.Transfer{
		TraceID:   uuidutil.Modify(output.TraceID, "compound_refund"),
		Opponents: []string{userID},
		Threshold: 1,
		AssetID:   output.AssetID,
		Amount:    output.UTXO.Amount,
		Memo:      "",
	}

	if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{transfer}); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("walletStore.CreateTransfers")
		return err
	}

	return nil
}

func (w *Payee) decodeMemo(memo string) []byte {
	if b, err := base64.StdEncoding.DecodeString(memo); err == nil {
		return b
	}

	if b, err := base64.URLEncoding.DecodeString(memo); err == nil {
		return b
	}

	return []byte(memo)
}
