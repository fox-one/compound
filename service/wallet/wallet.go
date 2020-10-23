package wallet

import (
	"compound/core"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/fox-one/mixin-sdk-go"
)

// New new wallet service
func New(c *mixin.Client, pin string) core.IWalletService {
	return &walletService{
		client: c,
		pin:    pin,
	}
}

type walletService struct {
	client *mixin.Client
	pin    string
}

func (s *walletService) HandleTransfer(ctx context.Context, transfer *core.Transfer) (*core.Snapshot, error) {
	input := &mixin.TransferInput{
		AssetID:    transfer.AssetID,
		OpponentID: transfer.OpponentID,
		Amount:     transfer.Amount,
		TraceID:    transfer.TraceID,
		Memo:       transfer.Memo,
	}

	snapshot, err := s.client.Transfer(ctx, input, s.pin)
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

	snapshots, err := s.client.ReadNetworkSnapshots(ctx, "", offset, "ASC", limit)
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
		ID:         snapshot.SnapshotID,
		TraceID:    snapshot.TraceID,
		CreatedAt:  snapshot.CreatedAt,
		UserID:     snapshot.UserID,
		OpponentID: snapshot.OpponentID,
		AssetID:    snapshot.AssetID,
		Amount:     snapshot.Amount,
		Memo:       snapshot.Memo,
	}
}

func (s *walletService) NewWallet(ctx context.Context, walletName, pin string) (*mixin.Keystore, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, "", err
	}

	_, keystore, err := s.client.CreateUser(ctx, privateKey, walletName)
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
