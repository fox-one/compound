package spentsync

import (
	"compound/core"
	"context"
	"crypto/md5"
	"errors"
	"io"
	"math/big"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store"
	"github.com/gofrs/uuid"
)

// SpentSync spent sync worker
type SpentSync struct {
	walletStore      core.WalletStore
	transactionStore core.TransactionStore
}

// New new spent sync worker
func New(
	walletStr core.WalletStore,
	transactionStr core.TransactionStore,
) *SpentSync {
	return &SpentSync{
		walletStore:      walletStr,
		transactionStore: transactionStr,
	}
}

// Run worker run
func (w *SpentSync) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "SpentSync")
	ctx = logger.WithContext(ctx, log)

	dur := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := w.run(ctx); err == nil {
				dur = 500 * time.Millisecond
			} else {
				dur = time.Second
			}
		}
	}
}

func (w *SpentSync) run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	const limit = 100
	transfers, err := w.walletStore.ListTransfers(ctx, core.TransferStatusHandled, limit)
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

	output, err := w.walletStore.FindSpentBy(ctx, transfer.AssetID, transfer.TraceID)
	if err != nil {
		if store.IsErrNotFound(err) {
			return nil
		}

		log.WithError(err).Errorln("wallets.ListSpentBy")
		return err
	}

	if output.State != mixin.UTXOStateSpent {
		log.Debugln("utxo is not spent, skip")
		return nil
	}

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
