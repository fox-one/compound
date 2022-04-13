package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

const (
	checkpointKey = "outputs_checkpoint"
	limit         = 500
)

type (
	// Payee payee worker
	Payee struct {
		system            *core.System
		dapp              *core.Wallet
		propertyStore     property.Store
		userStore         core.UserStore
		walletStore       core.WalletStore
		marketStore       core.IMarketStore
		supplyStore       core.ISupplyStore
		borrowStore       core.IBorrowStore
		proposalStore     core.ProposalStore
		transactionStore  core.TransactionStore
		oracleSignerStore core.OracleSignerStore
		walletz           core.WalletService
		proposalService   core.ProposalService
		accountService    core.IAccountService

		sysversion int64
	}
)

// NewPayee new payee
func NewPayee(
	system *core.System,
	dapp *core.Wallet,
	propertyStore property.Store,
	userStore core.UserStore,
	walletStore core.WalletStore,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	proposalStore core.ProposalStore,
	transactionStore core.TransactionStore,
	oracleSignerStr core.OracleSignerStore,
	walletz core.WalletService,
	proposalService core.ProposalService,
	accountService core.IAccountService,
) *Payee {

	payee := Payee{
		system:            system,
		dapp:              dapp,
		propertyStore:     propertyStore,
		userStore:         userStore,
		walletStore:       walletStore,
		marketStore:       marketStore,
		supplyStore:       supplyStore,
		borrowStore:       borrowStore,
		proposalStore:     proposalStore,
		transactionStore:  transactionStore,
		oracleSignerStore: oracleSignerStr,
		walletz:           walletz,
		proposalService:   proposalService,
		accountService:    accountService,
	}

	return &payee
}

// Run run worker
func (w *Payee) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "payee")
	ctx = logger.WithContext(ctx, log)

	dur := time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := w.run(ctx); err == nil {
				dur = 100 * time.Millisecond
			} else {
				dur = 500 * time.Millisecond
			}
		}
	}
}

func (w *Payee) run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "payee")

	if err := w.loadSysVersion(ctx); err != nil {
		return err
	}

	v, err := w.propertyStore.Get(ctx, checkpointKey)
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

	for _, output := range outputs {
		if err := w.handleOutput(ctx, output); err != nil {
			return err
		}

		if err := w.propertyStore.Save(ctx, checkpointKey, output.ID); err != nil {
			log.WithError(err).Errorln("property.Save", output.ID)
			return err
		}
	}

	return nil
}

func (w *Payee) handleOutput(ctx context.Context, output *core.Output) error {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"output":     output.TraceID,
		"asset":      output.AssetID,
		"amount":     output.Amount,
		"sysversion": w.sysversion,
		"sender":     output.Sender,
	})
	ctx = logger.WithContext(ctx, log)

	// handle price provided by dirtoracle
	{
		var e compound.Error
		if err := w.handlePriceEvent(ctx, output); err == nil {
			return nil
		} else if !errors.As(err, &e) {
			return err
		}
	}

	if output.Sender == "" {
		return nil
	}

	var (
		followID string
		err      error
	)
	switch w.sysversion {
	case 0, 1:
		err = w.handleOutputV1(ctx, output)

	case 2:
		err = w.handleOutputV2(ctx, output)

	default:
		followID, err = w.handleOutputV3(ctx, output)
	}

	var e compound.Error
	if !errors.As(err, &e) {
		return err
	}

	if compound.ShouldRefund(e.Flag) {
		memo := fmt.Sprintf(`{"f":"%s","m":"Rings operation failed: %s"}`, followID, e.Error())
		transfer := &core.Transfer{
			TraceID:   uuid.Modify(output.TraceID, memo),
			AssetID:   output.AssetID,
			Amount:    output.Amount,
			Memo:      base64.StdEncoding.EncodeToString([]byte(memo)),
			Threshold: 1,
			Opponents: []string{output.Sender},
		}

		if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{transfer}); err != nil {
			log.WithError(err).Errorln("wallets.CreateTransfers")
			return err
		}
	}
	return nil
}

func (w *Payee) transferOut(ctx context.Context, userID, followID, outputTraceID, assetID string, amount decimal.Decimal, transferAction *core.TransferAction) error {
	memoStr, e := transferAction.Format()
	if e != nil {
		return e
	}

	modifier := fmt.Sprintf("%s.%d", followID, transferAction.Source)
	transfer := core.Transfer{
		TraceID:   uuid.Modify(outputTraceID, modifier),
		Opponents: []string{userID},
		Threshold: 1,
		AssetID:   assetID,
		Amount:    amount,
		Memo:      memoStr,
	}

	if err := w.walletStore.CreateTransfers(ctx, []*core.Transfer{&transfer}); err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("wallets.CreateTransfers")
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
