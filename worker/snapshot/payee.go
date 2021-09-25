package snapshot

import (
	"compound/core"
	"compound/pkg/mtg"
	"compound/worker"
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const (
	checkpointKey = "outputs_checkpoint"
	limit         = 500
)

type (
	// Payee payee worker
	Payee struct {
		worker.TickWorker
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
		proposalService   core.ProposalService
		blockService      core.IBlockService
		marketService     core.IMarketService
		supplyService     core.ISupplyService
		borrowService     core.IBorrowService
		accountService    core.IAccountService
		allowListService  core.IAllowListService

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
	proposalService core.ProposalService,
	blockService core.IBlockService,
	marketSrv core.IMarketService,
	supplyService core.ISupplyService,
	borrowService core.IBorrowService,
	accountService core.IAccountService,
	allowListService core.IAllowListService) *Payee {

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
		proposalService:   proposalService,
		blockService:      blockService,
		marketService:     marketSrv,
		supplyService:     supplyService,
		borrowService:     borrowService,
		accountService:    accountService,
		allowListService:  allowListService,
	}

	return &payee
}

// Run run worker
func (w *Payee) Run(ctx context.Context) error {
	return w.StartTick(ctx, func(ctx context.Context) error {
		return w.onWork(ctx)
	})
}

func (w *Payee) onWork(ctx context.Context) error {
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

	for _, u := range outputs {
		if err := w.handleOutput(ctx, u); err != nil {
			return err
		}

		if err := w.propertyStore.Save(ctx, checkpointKey, u.ID); err != nil {
			log.WithError(err).Errorln("property.Save:", u.ID)
			return err
		}
	}

	return nil
}

func (w *Payee) handleOutput(ctx context.Context, output *core.Output) error {
	log := logger.FromContext(ctx).
		WithField("output", output.TraceID).
		WithField("sysversion", w.sysversion)
	ctx = logger.WithContext(ctx, log)

	businessData := w.decodeMemo(output.Memo)

	if w.sysversion < 1 {
		// handle member proposal action
		if member, action, body, err := core.DecodeMemberActionV0(businessData, w.system.Members); err == nil {
			return w.handleProposalActionV0(ctx, output, member, action, body)
		}
	} else {
		// handle member proposal action
		if member, body, err := core.DecodeMemberActionV1(businessData, w.system.Members); err == nil {
			return w.handleMemberAction(ctx, output, member.ClientID, body)
		}
	}

	// handle price provided by dirtoracle
	if priceData, err := w.decodePriceTransaction(ctx, businessData); err == nil {
		return w.handlePriceEvent(ctx, output, priceData)
	}

	// decode user action
	actionType, body, err := core.DecodeTransactionAction(w.system.PrivateKey, businessData)
	if err != nil {
		log.WithError(err).Errorln("DecodeTransactionAction error")
		return nil
	}

	// transaction trace id as order id, different from output trace id
	var followID uuid.UUID
	body, err = mtg.Scan(body, &followID)
	if err != nil {
		log.WithError(err).Errorln("scan userID and followID error")
		return nil
	}

	if output.Sender == "" {
		return nil
	}

	//upsert user
	user := core.User{
		UserID:  output.Sender,
		Address: core.BuildUserAddress(output.Sender),
	}
	if err = w.userStore.Save(ctx, &user); err != nil {
		return err
	}

	// handle user action
	return w.handleUserAction(ctx, output, actionType, output.Sender, followID.String(), body)
}

func (w *Payee) handleUserAction(ctx context.Context, output *core.Output, actionType core.ActionType, userID, followID string, body []byte) error {
	switch actionType {
	case core.ActionTypeSupply:
		return w.handleSupplyEvent(ctx, output, userID, followID, body)
	case core.ActionTypeBorrow:
		return w.handleBorrowEvent(ctx, output, userID, followID, body)
	case core.ActionTypeRedeem:
		return w.handleRedeemEvent(ctx, output, userID, followID, body)
	case core.ActionTypeRepay:
		return w.handleRepayEvent(ctx, output, userID, followID, body)
	case core.ActionTypePledge:
		return w.handlePledgeEvent(ctx, output, userID, followID, body)
	case core.ActionTypeUnpledge:
		return w.handleUnpledgeEvent(ctx, output, userID, followID, body)
	case core.ActionTypeQuickPledge:
		return w.handleQuickPledgeEvent(ctx, output, userID, followID, body)
	case core.ActionTypeQuickBorrow:
		return w.handleQuickBorrowEvent(ctx, output, userID, followID, body)
	case core.ActionTypeQuickRedeem:
		return w.handleQuickRedeemEvent(ctx, output, userID, followID, body)
	case core.ActionTypeLiquidate:
		return w.handleLiquidationEvent(ctx, output, userID, followID, body)
	default:
		return w.handleRefundEvent(ctx, output, userID, followID, core.ActionTypeRefundTransfer, core.ErrUnknown)
	}
}

func (w *Payee) transferOut(ctx context.Context, userID, followID, outputTraceID, assetID string, amount decimal.Decimal, transferAction *core.TransferAction) error {
	memoStr, e := transferAction.Format()
	if e != nil {
		return e
	}

	modifier := fmt.Sprintf("%s.%d", followID, transferAction.Source)
	transfer := core.Transfer{
		TraceID:   uuidutil.Modify(outputTraceID, modifier),
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
