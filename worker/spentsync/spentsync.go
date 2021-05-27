package spentsync

import (
	"compound/core"
	"compound/worker"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/gofrs/uuid"
)

// SpentSync spent sync worker
type SpentSync struct {
	worker.TickWorker
	db               *db.DB
	walletStore      core.WalletStore
	transactionStore core.TransactionStore
}

// New new spent sync worker
func New(
	db *db.DB,
	walletStr core.WalletStore,
	transactionStr core.TransactionStore,
) *SpentSync {
	return &SpentSync{
		db:               db,
		walletStore:      walletStr,
		transactionStore: transactionStr,
	}
}

// Run worker run
func (w *SpentSync) Run(ctx context.Context) error {
	return w.StartTick(ctx, func(ctx context.Context) error {
		return w.onWork(ctx)
	})
}

func (w *SpentSync) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx)

	transfers, err := w.walletStore.ListNotPassedTransfers(ctx)
	if err != nil {
		log.WithError(err).Errorln("wallets.ListNotPassedTransfers")
		return err
	}

	if len(transfers) == 0 {
		return errors.New("EOF")
	}

	for _, transfer := range transfers {
		err = w.handleTransfer(ctx, transfer)
		if err != nil {
			continue
		}
	}

	return nil
}

func (w *SpentSync) handleTransfer(ctx context.Context, transfer *core.Transfer) error {
	log := logger.FromContext(ctx).WithField("trace", transfer.TraceID)

	log.Debugf("handle transfer")

	outputs, err := w.walletStore.ListSpentBy(ctx, transfer.AssetID, transfer.TraceID)
	if err != nil {
		log.WithError(err).Errorln("wallets.ListSpentBy")
		return err
	}

	if len(outputs) == 0 {
		log.Debugln("no outputs spent, skip")
		return nil
	}

	output := outputs[0]
	if output.State != mixin.UTXOStateSpent {
		log.Debugln("utxo is not spent, skip")
		return nil
	}

	fmt.Println("ID:", output.ID, ":trace:", output.TraceID, "asset", output.AssetID, ":amount:", output.Amount, ":memo:", output.Memo, ":data:", output.Data)

	signedTx := output.UTXO.SignedTx
	if signedTx == "" {
		log.Debugln("signed tx is empty, skip")
		return nil
	}

	//add transaction
	snapshotTraceID, err := w.snapshotTraceID(ctx, signedTx)
	if err != nil {
		log.WithError(err).Errorln("get snapshot trace id error")
		return nil
	}
	transaction, err := core.BuildTransactionFromTransfer(ctx, transfer, snapshotTraceID)
	if err != nil {
		return err
	}
	if err = w.transactionStore.Create(ctx, transaction); err != nil {
		log.WithError(err).Errorln("create transaction error")
		return err
	}

	//update transfer
	transfer.Passed = true
	if err := w.walletStore.UpdateTransfer(ctx, transfer); err != nil {
		log.WithError(err).Errorln("wallets.UpdateTransfer")
		return err
	}

	return nil
}

func (w *SpentSync) snapshotTraceID(ctx context.Context, signedTx string) (string, error) {
	log := logger.FromContext(ctx)

	tx, err := mixin.TransactionFromRaw(signedTx)
	if err != nil {
		log.WithError(err).Debugln("decode transaction from raw tx failed")
		return "", err
	}

	hash, err := tx.TransactionHash()
	if err != nil {
		return "", err
	}

	traceID, err := w.mixinRawTransactionTraceID(hash.String(), 0)
	if err != nil {
		return "", err
	}
	return traceID, nil
}

func (w *SpentSync) mixinRawTransactionTraceID(hash string, index uint8) (string, error) {
	h := md5.New()
	_, err := io.WriteString(h, hash)
	if err != nil {
		return "", err
	}
	b := new(big.Int).SetInt64(int64(index))
	_, err = h.Write(b.Bytes())
	if err != nil {
		return "", err
	}
	s := h.Sum(nil)
	s[6] = (s[6] & 0x0f) | 0x30
	s[8] = (s[8] & 0x3f) | 0x80
	sid, err := uuid.FromBytes(s)
	if err != nil {
		return "", err
	}

	return sid.String(), nil
}
