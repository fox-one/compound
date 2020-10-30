package snapshot

import (
	"compound/core"
	"compound/worker"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/bluele/gcache"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/property"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

// Worker snapshot worker
type Worker struct {
	worker.BaseJob
	config        *core.Config
	dapp          *mixin.Client
	property      property.Store
	walletService core.IWalletService
	blockService  core.IBlockService
	priceService  core.IPriceOracleService
	marketService core.IMarketService
	snapshotCache gcache.Cache
}

const (
	checkPointKey = "compound_snapshot_checkpoint"
	limit         = 500
)

// New new snapshot worker
func New(
	config *core.Config,
	dapp *mixin.Client,
	property property.Store,
	walletService core.IWalletService,
	priceSrv core.IPriceOracleService,
	blockService core.IBlockService,
	marketSrv core.IMarketService,
) *Worker {
	job := Worker{
		config:        config,
		dapp:          dapp,
		property:      property,
		walletService: walletService,
		blockService:  blockService,
		priceService:  priceSrv,
		marketService: marketSrv,
		snapshotCache: gcache.New(limit).LRU().Build(),
	}

	l, _ := time.LoadLocation(job.config.App.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 1s"
	job.Cron.AddFunc(spec, job.Run)
	job.OnWork = func() error {
		return job.onWork(context.Background())
	}

	return &job
}

func (w *Worker) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx)
	checkPoint, err := w.property.Get(ctx, checkPointKey)
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

		// if w.snapshotCache.Has(snapshot.ID) {
		// 	continue
		// }

		if err := w.handleSnapshot(ctx, snapshot); err != nil {
			return err
		}

		// w.snapshotCache.Set(snapshot.ID, nil)
	}

	if checkPoint.String() != next {
		if err := w.property.Save(ctx, checkPointKey, next); err != nil {
			log.WithError(err).Errorf("update property error: %s", checkPointKey)
			return err
		}
	}

	return nil
}

func (w *Worker) handleSnapshot(ctx context.Context, snapshot *core.Snapshot) error {
	if snapshot.UserID == w.config.BlockWallet.ClientID {
		return w.handleBlockEvent(ctx, snapshot)
	} else {

	}
	return nil
}

func (w *Worker) handleBlockEvent(ctx context.Context, snapshot *core.Snapshot) error {
	if snapshot.AssetID != w.config.App.BlockAssetID {
		return nil
	}

	log := logger.FromContext(ctx).WithField("worker", "snapshot")

	blockMemo, err := w.blockService.ParseBlockMemo(ctx, snapshot.Memo)
	if err != nil {
		log.Errorln("parse block memo error:", err)
		return nil
	}

	block, err := strconv.ParseInt(blockMemo[core.BlockMemoKeyBlock], 10, 64)
	if err != nil {
		return nil
	}

	service := blockMemo[core.BlockMemoKeyService]
	if service == core.MemoServicePrice {
		// cache price

		symbol := blockMemo[core.BlockMemoKeySymbol]
		price, err := decimal.NewFromString(blockMemo[core.BlockMemoKeyPrice])
		if err != nil {
			return nil
		}

		w.priceService.Save(ctx, symbol, price, block)
	} else if service == core.MemoServiceMarket {
		symbol := blockMemo[core.BlockMemoKeySymbol]
	
		//utilization rate
		utilizationRate, err := decimal.NewFromString(blockMemo[core.BlockMemoKeyUtilizationRate])
		if err != nil {
			return nil
		}

		w.marketService.SaveUtilizationRate(ctx, symbol, utilizationRate, block)

		// borrow rate
		borrowRate, err := decimal.NewFromString(blockMemo[core.BlockMemoKeyBorrowRate])
		if err != nil {
			return nil
		}

		w.marketService.SaveBorrowRatePerBlock(ctx, symbol, borrowRate, block)

		// supply rate
		supplyRate, err := decimal.NewFromString(blockMemo[core.BlockMemoKeySupplyRate])
		if err != nil {
			return nil
		}
		w.marketService.SaveSupplyRatePerBlock(ctx, symbol, supplyRate, block)
	}

	// cache market

	//market
	//calculate utilization rate
	//calculate exchange rate
	//calculate borrow rate
	//calculate supply rate

	//market
	//calculate borrow interest
	//calculate supply interest

	//market
	//calcutate reserve

	//account
	//scan account liquidity

	return nil
}
