package assigner

import (
	"compound/core"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

var (
	errOutputMergeRequired = errors.New("output merge required")
)

func New(
	wallets core.WalletStore,
	system *core.System,
) *Assigner {
	return &Assigner{
		wallets: wallets,
		system:  system,
	}
}

type Assigner struct {
	wallets core.WalletStore
	system  *core.System
}

func (w *Assigner) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "assigner")
	ctx = logger.WithContext(ctx, log)

	dur := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := w.run(ctx); err == nil {
				dur = 300 * time.Millisecond
			} else {
				dur = 500 * time.Millisecond
			}
		}
	}
}

func (w *Assigner) run(ctx context.Context) error {
	const limit = 100
	transfers, err := w.wallets.ListTransfers(ctx, core.TransferStatusPending, limit)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.ListTransfers")
		return err
	}

	aborted := map[string]bool{}
	for _, transfer := range transfers {
		if aborted[transfer.AssetID] {
			continue
		}

		if err := w.handleTransfer(ctx, transfer); err != nil {
			aborted[transfer.AssetID] = true
		}
	}

	if len(aborted) > 0 {
		return errors.New("aborted")
	}

	return nil
}

func (w *Assigner) handleTransfer(ctx context.Context, transfer *core.Transfer) error {
	log := logger.FromContext(ctx).WithField("transfer", transfer.TraceID)

	const limit = 32
	outputs, err := w.wallets.ListUnspent(ctx, transfer.AssetID, limit)
	if err != nil {
		log.WithError(err).Errorln("wallets.ListUnspent")
		return err
	}

	var (
		idx    int
		sum    decimal.Decimal
		traces []string
	)

	// spent all utxo if all utxo is older than the transfer
	wipe := len(outputs) == limit && outputs[limit-1].ID <= transfer.Version

	for _, output := range outputs {
		sum = sum.Add(output.Amount)
		traces = append(traces, output.TraceID)
		idx += 1

		if wipe {
			continue
		}

		if sum.GreaterThanOrEqual(transfer.Amount) {
			break
		}
	}

	outputs = outputs[:idx]

	if sum.LessThan(transfer.Amount) {
		// merge outputs
		if len(outputs) == limit {
			traceID := uuid.Modify(transfer.TraceID, mixin.HashMembers(traces))
			merge := &core.Transfer{
				TraceID:   traceID,
				Version:   transfer.Version,
				AssetID:   transfer.AssetID,
				Amount:    sum,
				Opponents: w.system.MemberIDs,
				Threshold: w.system.Threshold,
				Memo:      fmt.Sprintf("merge for %s", transfer.TraceID),
			}

			if err := w.commit(ctx, outputs, merge); err != nil {
				return err
			}

			return errOutputMergeRequired
		}

		err := errors.New("insufficient balance")
		log.WithError(err).Errorln("handle transfer", transfer.ID)
		return err
	}

	return w.commit(ctx, outputs, transfer)
}

func (w *Assigner) commit(ctx context.Context, outputs []*core.Output, transfer *core.Transfer) error {
	if err := w.wallets.Assign(ctx, outputs, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.Assign")
		return err
	}

	return nil
}
