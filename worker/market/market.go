package market

import (
	"compound/core"
	"compound/pkg/id"
	"compound/worker"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

//Worker market worker
type Worker struct {
	worker.BaseJob
	MainWallet    *core.Wallet
	BlockWallet   *core.Wallet
	Config        *core.Config
	MarketStore   core.IMarketStore
	BlockService  core.IBlockService
	PriceService  core.IPriceOracleService
	MarketService core.IMarketService
	WalletService core.IWalletService
}

// New new market worker
func New(mainWallet *core.Wallet, blockWallet *core.Wallet, cfg *core.Config, marketStore core.IMarketStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService, walletSrv core.IWalletService) *Worker {
	job := Worker{
		MainWallet:    mainWallet,
		BlockWallet:   blockWallet,
		Config:        cfg,
		MarketStore:   marketStore,
		BlockService:  blockSrv,
		PriceService:  priceSrv,
		WalletService: walletSrv,
	}

	l, _ := time.LoadLocation(job.Config.App.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 10ms"
	job.Cron.AddFunc(spec, job.Run)
	job.OnWork = func() error {
		return job.onWork(context.Background())
	}

	return &job
}

func (w *Worker) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "market")

	markets, err := w.MarketStore.All(ctx)
	if err != nil {
		log.Errorln("fetch all markets error:", err)
		return err
	}

	wg := sync.WaitGroup{}
	for _, m := range markets {
		wg.Add(1)
		go func(market *core.Market) {
			defer wg.Done()
			// do real work
			w.checkAndPushMarketOnChain(ctx, market)
		}(m)
	}

	wg.Wait()

	return nil
}

func (w *Worker) checkAndPushMarketOnChain(ctx context.Context, market *core.Market) error {
	log := logger.FromContext(ctx).WithField("worker", "market")

	currentBlock, err := w.BlockService.CurrentBlock(ctx)
	if err != nil {
		log.Errorln(err)
		return err
	}

	str := fmt.Sprintf("foxone-compound-market-%s-%d", market.Symbol, currentBlock)
	traceID := id.UUIDFromString(str)
	transferInput := mixin.TransferInput{
		AssetID:    w.Config.App.BlockAssetID,
		OpponentID: w.MainWallet.Client.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    traceID,
	}

	if !w.WalletService.VerifyPayment(ctx, &transferInput) {
		//create new block
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceMarket
		memo[core.ActionKeySymbol] = market.Symbol
		memo[core.ActionKeyBlock] = strconv.FormatInt(currentBlock, 10)

		// utilization rate
		uRate, err := w.MarketService.CurUtilizationRate(ctx, market)
		if err != nil {
			log.Errorln("get utilization rate error:", err)
			return err
		}

		memo[core.ActionKeyUtilizationRate] = uRate.Truncate(4).String()

		// borrow rate
		bRate, err := w.MarketService.CurBorrowRatePerBlock(ctx, market)
		if err != nil {
			log.Errorln("get borrow rate per block error:", err)
			return err
		}

		memo[core.ActionKeyBorrowRate] = bRate.Truncate(16).String()

		// supply rate
		sRate, err := w.MarketService.CurSupplyRatePerBlock(ctx, market)
		if err != nil {
			log.Errorln("get supply rate per block error:", err)
			return err
		}

		memo[core.ActionKeySupplyRate] = sRate.Truncate(16).String()

		// format memo to string
		memoStr, err := memo.Format()
		if err != nil {
			log.Errorln("new block memo error:", err)
			return err
		}

		transferInput.Memo = memoStr
		_, err = w.BlockWallet.Client.Transfer(ctx, &transferInput, w.BlockWallet.Pin)

		if err != nil {
			log.Errorln("transfer new block error:", err)
			return err
		}
	}

	return nil
}
