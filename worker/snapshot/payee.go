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
	"github.com/fox-one/pkg/store/db"
	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/gofrs/uuid"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

const (
	checkpointKey = "outputs_checkpoint"
	limit         = 500
)

// Payee payee worker
type Payee struct {
	worker.BaseJob
	db              *db.DB
	system          *core.System
	dapp            *core.Wallet
	propertyStore   property.Store
	walletStore     core.WalletStore
	marketStore     core.IMarketStore
	supplyStore     core.ISupplyStore
	borrowStore     core.IBorrowStore
	proposalStore   core.ProposalStore
	proposalService core.ProposalService
	blockService    core.IBlockService
	priceService    core.IPriceOracleService
	marketService   core.IMarketService
	supplyService   core.ISupplyService
	borrowService   core.IBorrowService
	accountService  core.IAccountService
}

// NewPayee new payee
func NewPayee(location string,
	db *db.DB,
	system *core.System,
	dapp *core.Wallet,
	propertyStore property.Store,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	proposalStore core.ProposalStore,
	proposalService core.ProposalService,
	priceSrv core.IPriceOracleService,
	blockService core.IBlockService,
	marketSrv core.IMarketService,
	supplyService core.ISupplyService,
	borrowService core.IBorrowService,
	accountService core.IAccountService) *Payee {
	payee := Payee{
		db:              db,
		system:          system,
		dapp:            dapp,
		propertyStore:   propertyStore,
		marketStore:     marketStore,
		supplyStore:     supplyStore,
		borrowStore:     borrowStore,
		proposalStore:   proposalStore,
		proposalService: proposalService,
		priceService:    priceSrv,
		blockService:    blockService,
		marketService:   marketSrv,
		supplyService:   supplyService,
		borrowService:   borrowService,
		accountService:  accountService,
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
	log := logger.FromContext(ctx).WithField("output", output.TraceID)
	ctx = logger.WithContext(ctx, log)

	message := w.decodeMemo(output.Memo)

	// handle member vote action
	if member, body, err := core.DecodeMemberProposalTransactionAction(message, w.system.Members); err == nil {
		return w.handleProposalAction(ctx, output, member, body)
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

	return w.handleUserAction(ctx, output, actionType, userID.String(), followID.String(), body)
}

func (w *Payee) handleProposalAction(ctx context.Context, output *core.Output, member *core.Member, body []byte) error {
	log := logger.FromContext(ctx)

	var traceID uuid.UUID
	var actionType int

	body, err := mtg.Scan(body, &traceID, &actionType)
	if err != nil {
		log.WithError(err).Debugln("scan proposal trace & action failed")
		return nil
	}

	if core.ActionType(actionType) == core.ActionTypeProposalVote {
		return w.handleVoteProposalEvent(ctx, output, member, traceID.String())
	}

	return w.handleCreateProposalEvent(ctx, output, member, core.ActionType(actionType), traceID.String(), body)
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
		return w.handleReplayEvent(ctx, output, userID, followID, body)
	case core.ActionTypePledge:
		return w.handlePledgeEvent(ctx, output, userID, followID, body)
	case core.ActionTypeUnpledge:
		return w.handleUnpledgeEvent(ctx, output, userID, followID, body)
	case core.ActionTypeSeizeToken:
		return w.handleSeizeTokenEvent(ctx, output, userID, followID, body)
	default:
		return w.handleRefundEvent(ctx, output, userID, followID, core.ErrUnknown, "")
	}

}

func (w *Payee) transferOut(ctx context.Context, userID, followID, outputTraceID, assetID string, amount decimal.Decimal, transferAction *core.TransferAction) error {
	memoStr, e := transferAction.Format()
	if e != nil {
		return e
	}

	transfer := core.Transfer{
		TraceID:   uuidutil.Modify(outputTraceID, followID),
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
