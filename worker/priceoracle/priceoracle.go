package priceoracle

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

//Worker block worker
type Worker struct {
	worker.BaseJob
	MainWallet         *core.Wallet
	BlockWallet        *core.Wallet
	Config             *core.Config
	MarketStore        core.IMarketStore
	BlockService       core.IBlockService
	PriceOracleService core.IPriceOracleService
	WalletService      core.IWalletService
}

// New new block worker
func New(mainWallet *core.Wallet, blockWallet *core.Wallet, cfg *core.Config, marketStore core.IMarketStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService, walletSrv core.IWalletService) *Worker {
	job := Worker{
		MainWallet:         mainWallet,
		BlockWallet:        blockWallet,
		Config:             cfg,
		MarketStore:        marketStore,
		BlockService:       blockSrv,
		PriceOracleService: priceSrv,
		WalletService:      walletSrv,
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
	log := logger.FromContext(ctx).WithField("worker", "priceoracle")

	markets, err := w.MarketStore.All(ctx)
	if err != nil {
		log.Errorln("fetch all markets error:", err)
		return err
	}

	if len(markets) <= 0 {
		log.Infoln("no market found!!!")
		return nil
	}

	wg := sync.WaitGroup{}
	for _, m := range markets {
		wg.Add(1)
		go func(market *core.Market) {
			defer wg.Done()
			// do real work
			ticker, e := w.PriceOracleService.PullPriceTicker(ctx, market.Symbol, time.Now())
			if e != nil {
				log.Errorln("pull price ticker error:", e)
				return
			}
			if ticker.Price.LessThanOrEqual(decimal.Zero) {
				log.Errorln("invalid ticker price:", ticker.Symbol, ":", ticker.Price)
				return
			}

			w.checkAndPushPriceOnChain(ctx, market, ticker)
		}(m)
	}

	wg.Wait()

	return nil
}

func (w *Worker) checkAndPushPriceOnChain(ctx context.Context, market *core.Market, ticker *core.PriceTicker) error {
	log := logger.FromContext(ctx).WithField("worker", "priceoracle")

	blockNum, err := w.BlockService.GetBlock(ctx, time.Now())
	if err != nil {
		log.Errorln(err)
		return err
	}

	str := fmt.Sprintf("foxone-compound-price-%s-%d", market.Symbol, blockNum)
	traceID := id.UUIDFromString(str)
	transferInput := mixin.TransferInput{
		AssetID:    w.Config.App.GasAssetID,
		OpponentID: w.MainWallet.Client.ClientID,
		Amount:     core.GasCost,
		TraceID:    traceID,
	}

	if !w.WalletService.VerifyPayment(ctx, &transferInput) {
		//create new block
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServicePrice
		memo[core.ActionKeyBlock] = strconv.FormatInt(blockNum, 10)
		memo[core.ActionKeySymbol] = market.Symbol
		memo[core.ActionKeyPrice] = ticker.Price.Truncate(8).String()
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
