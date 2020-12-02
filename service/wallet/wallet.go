package wallet

import (
	"compound/core"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// New new wallet service
func New(mainWallet *core.Wallet) core.IWalletService {
	return &walletService{
		MainWallet: mainWallet,
	}
}

type walletService struct {
	MainWallet *core.Wallet
}

func (s *walletService) HandleTransfer(ctx context.Context, transfer *core.Transfer) (*core.Snapshot, error) {
	input := &mixin.TransferInput{
		AssetID:    transfer.AssetID,
		OpponentID: transfer.OpponentID,
		Amount:     transfer.Amount,
		TraceID:    transfer.TraceID,
		Memo:       transfer.Memo,
	}

	snapshot, err := s.MainWallet.Client.Transfer(ctx, input, s.MainWallet.Pin)
	if err != nil {
		return nil, err
	}

	return convertSnapshot(snapshot), nil
}

func (s *walletService) PullSnapshots(ctx context.Context, cursor string, limit int) ([]*core.Snapshot, string, error) {
	offset, err := time.Parse(time.RFC3339Nano, cursor)
	if err != nil {
		offset = time.Now().UTC()
	}

	snapshots, err := s.MainWallet.Client.ReadNetworkSnapshots(ctx, "", offset, "ASC", limit)
	if err != nil {
		return nil, "", err
	}

	out := make([]*core.Snapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		out = append(out, convertSnapshot(snapshot))
		offset = snapshot.CreatedAt
	}

	return out, offset.Format(time.RFC3339Nano), nil
}

func convertSnapshot(snapshot *mixin.Snapshot) *core.Snapshot {
	return &core.Snapshot{
		SnapshotID: snapshot.SnapshotID,
		TraceID:    snapshot.TraceID,
		UserID:     snapshot.UserID,
		OpponentID: snapshot.OpponentID,
		AssetID:    snapshot.AssetID,
		Amount:     snapshot.Amount,
		Memo:       snapshot.Memo,
		CreatedAt:  snapshot.CreatedAt,
	}
}

func (s *walletService) NewWallet(ctx context.Context, walletName, pin string) (*mixin.Keystore, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, "", err
	}

	_, keystore, err := s.MainWallet.Client.CreateUser(ctx, privateKey, walletName)
	if err != nil {
		return nil, "", err
	}

	newClient, err := mixin.NewFromKeystore(keystore)
	if err != nil {
		return nil, "", err
	}

	err = newClient.ModifyPin(ctx, "", pin)
	if err != nil {
		return nil, "", err
	}

	return keystore, pin, nil
}

// PaySchemaURL build pay schema url
func (s *walletService) PaySchemaURL(amount decimal.Decimal, asset, recipient, trace, memo string) (string, error) {
	if amount.LessThanOrEqual(decimal.Zero) || asset == "" || recipient == "" || trace == "" {
		return "", errors.New("invalid paramaters")
	}

	return fmt.Sprintf("mixin://pay?amount=%s&asset=%s&recipient=%s&trace=%s&memo=%s", amount.String(), asset, recipient, trace, memo), nil
}

func (s *walletService) VerifyPayment(ctx context.Context, input *mixin.TransferInput) bool {
	log := logger.FromContext(ctx)

	payment, err := s.MainWallet.Client.VerifyPayment(ctx, *input)
	if err != nil {
		log.Errorln("verifypayment error:", err)
		return false
	}

	return payment.Status == "paid"
}
