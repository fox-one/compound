package snapshot

import (
	"compound/core"
	"compound/worker"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
	"github.com/fox-one/pkg/store/db"
	"github.com/robfig/cron/v3"
)

// Worker snapshot worker
type Worker struct {
	worker.BaseJob
	config                   *core.Config
	db                       *db.DB
	mainWallet               *core.Wallet
	blockWallet              *core.Wallet
	propertyStore            property.Store
	transferStore            core.ITransferStore
	marketStore              core.IMarketStore
	supplyStore              core.ISupplyStore
	borrowStore              core.IBorrowStore
	walletService            core.IWalletService
	blockService             core.IBlockService
	priceService             core.IPriceOracleService
	marketService            core.IMarketService
	supplyService            core.ISupplyService
	borrowService            core.IBorrowService
	accountService           core.IAccountService
	transactionEventHandlers map[string]handleTransactionEventFunc
}

const (
	checkPointKey = "compound_snapshot_checkpoint"
	limit         = 500
)

// New new snapshot worker
func New(
	config *core.Config,
	db *db.DB,
	mainWallet *core.Wallet,
	blockWallet *core.Wallet,
	propertyStore property.Store,
	transferStore core.ITransferStore,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	walletService core.IWalletService,
	priceSrv core.IPriceOracleService,
	blockService core.IBlockService,
	marketSrv core.IMarketService,
	supplyService core.ISupplyService,
	borrowService core.IBorrowService,
	accountService core.IAccountService,
) *Worker {
	//add transaction event handle
	handlers := make(map[string]handleTransactionEventFunc)
	handlers[core.ActionServicePrice] = handlePriceEvent
	handlers[core.ActionServiceSupply] = handleSupplyEvent
	handlers[core.ActionServiceRedeem] = handleSupplyRedeemEvent
	handlers[core.ActionServicePledge] = handlePledgeEvent
	handlers[core.ActionServiceUnpledge] = handleUnpledgeEvent
	handlers[core.ActionServiceBorrow] = handleBorrowEvent
	handlers[core.ActionServiceRepay] = handleBorrowRepayEvent
	handlers[core.ActionServiceSeizeToken] = handleSeizeTokenEvent
	handlers[core.ActionServiceAddMarket] = handleAddMarketEvent
	handlers[core.ActionServiceUpdateMarket] = handleUpdateMarketEvent

	job := Worker{
		config:                   config,
		db:                       db,
		mainWallet:               mainWallet,
		blockWallet:              blockWallet,
		propertyStore:            propertyStore,
		transferStore:            transferStore,
		marketStore:              marketStore,
		supplyStore:              supplyStore,
		borrowStore:              borrowStore,
		walletService:            walletService,
		blockService:             blockService,
		priceService:             priceSrv,
		marketService:            marketSrv,
		supplyService:            supplyService,
		borrowService:            borrowService,
		accountService:           accountService,
		transactionEventHandlers: handlers,
	}

	l, _ := time.LoadLocation(job.config.App.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 100ms"
	job.Cron.AddFunc(spec, job.Run)
	job.OnWork = func() error {
		return job.onWork(context.Background())
	}

	return &job
}

func (w *Worker) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx)
	checkPoint, err := w.propertyStore.Get(ctx, checkPointKey)
	if err != nil {
		log.WithError(err).Errorf("read property error: %s", checkPointKey)
		return err
	}

	snapshots, next, err := w.walletService.PullSnapshots(ctx, checkPoint.String(), limit)
	if err != nil {
		log.WithError(err).Error("pull snapshots error")
		return err
	}

	if len(snapshots) == 0 {
		return errors.New("no more snapshots")
	}

	for _, snapshot := range snapshots {
		if snapshot.UserID == "" {
			continue
		}

		if err := w.handleSnapshot(ctx, snapshot); err != nil {
			return err
		}
	}

	if checkPoint.String() != next {
		if err := w.propertyStore.Save(ctx, checkPointKey, next); err != nil {
			log.WithError(err).Errorf("update property error: %s", checkPointKey)
			return err
		}
	}

	return nil
}

func (w *Worker) handleSnapshot(ctx context.Context, snapshot *core.Snapshot) error {
	if snapshot.UserID == w.mainWallet.Client.ClientID {
		// main wallet
		var action core.Action
		e := json.Unmarshal([]byte(snapshot.Memo), &action)
		if e != nil {
			return nil
		}

		service := action[core.ActionKeyService]

		handlerTransactionEvent, found := w.transactionEventHandlers[service]
		if found {
			return handlerTransactionEvent(ctx, w, action, snapshot)
		}

		return handleRefundEvent(ctx, w, action, snapshot, core.ErrUnknown)
	}
	return nil
}

type handleTransactionEventFunc func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error
