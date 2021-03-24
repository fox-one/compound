package cashier

import (
	"context"
	"errors"
	"fmt"

	"compound/core"
	"compound/worker"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

// Cashier cashier
//
// use output to spend
type Cashier struct {
	worker.TickWorker
	walletStore   core.WalletStore
	walletService core.WalletService
	system        *core.System
}

// New new cashier
func New(
	walletStr core.WalletStore,
	walletSrv core.WalletService,
	system *core.System,
) *Cashier {
	cashier := Cashier{
		walletStore:   walletStr,
		walletService: walletSrv,
		system:        system,
	}

	return &cashier
}

// Run run worker
func (w *Cashier) Run(ctx context.Context) error {
	return w.StartTick(ctx, func(ctx context.Context) error {
		return w.onWork(ctx)
	})
}

func (w *Cashier) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "cashier")

	transfers, err := w.walletStore.ListPendingTransfers(ctx)
	if err != nil {
		log.WithError(err).Errorln("list transfers")
		return err
	}

	if len(transfers) == 0 {
		return errors.New("EOF")
	}

	for _, transfer := range transfers {
		_ = w.handleTransfer(ctx, transfer)
	}

	return nil
}

func (w *Cashier) handleTransfer(ctx context.Context, transfer *core.Transfer) error {
	log := logger.FromContext(ctx)

	const limit = 64
	outputs, err := w.walletStore.ListUnspent(ctx, transfer.AssetID, limit)
	if err != nil {
		log.WithError(err).Errorln("wallets.ListUnspent")
		return err
	}

	var (
		idx    int
		sum    decimal.Decimal
		traces []string
	)

	for _, output := range outputs {
		sum = sum.Add(output.Amount)
		traces = append(traces, output.TraceID)
		idx++

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
				AssetID:   transfer.AssetID,
				Amount:    sum,
				Opponents: w.system.MemberIDs(),
				Threshold: w.system.Threshold,
				Memo:      fmt.Sprintf("merge for %s", transfer.TraceID),
			}

			return w.spent(ctx, outputs, merge)
		}

		err := errors.New("insufficient balance")
		log.WithError(err).Errorln("handle transfer", transfer.ID)
		return err
	}

	return w.spent(ctx, outputs, transfer)
}

func (w *Cashier) spent(ctx context.Context, outputs []*core.Output, transfer *core.Transfer) error {
	if tx, err := w.walletService.Spent(ctx, outputs, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("walletz.Spent")
		return err
	} else if tx != nil {
		// 签名收集完成，需要提交至主网
		// 此时将该上链 tx 存储至数据库，等待 tx sender worker 完成上链
		if err := w.walletStore.CreateRawTransaction(ctx, tx); err != nil {
			logger.FromContext(ctx).WithError(err).Errorln("wallets.CreateRawTransaction")
			return err
		}
	}

	if err := w.walletStore.Spent(ctx, outputs, transfer); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.Spent")
		return err
	}

	return nil
}
