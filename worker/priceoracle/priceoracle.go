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
	MixinDapp          *mixin.Client
	BlockWallet        *mixin.Client
	Config             *core.Config
	MarketStore        core.IMarketStore
	BlockService       core.IBlockService
	PriceOracleService core.IPriceOracleService
}

// New new block worker
func New(dapp *mixin.Client, blockWallet *mixin.Client, cfg *core.Config, marketStore core.IMarketStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService) *Worker {
	job := Worker{
		MixinDapp:          dapp,
		BlockWallet:        blockWallet,
		Config:             cfg,
		MarketStore:        marketStore,
		BlockService:       blockSrv,
		PriceOracleService: priceSrv,
	}

	l, _ := time.LoadLocation(job.Config.App.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 250ms"
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

	currentBlock, err := w.BlockService.CurrentBlock(ctx)
	if err != nil {
		log.Errorln(err)
		return err
	}

	str := fmt.Sprintf("foxone-compound-price-%s-%d", market.Symbol, currentBlock)
	traceID := id.UUIDFromString(str)
	transferInput := mixin.TransferInput{
		AssetID:    w.Config.App.BlockAssetID,
		OpponentID: w.Config.Mixin.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    traceID,
	}
	payment, err := w.MixinDapp.VerifyPayment(ctx, transferInput)
	if err != nil {
		log.Errorln("verifypayment error:", err)
		return err
	}

	if payment.Status == "paid" {
		log.Infoln("transation exists")
	} else {
		//create new block
		memo := make(core.BlockMemo)
		memo[core.BlockMemoKeyService] = core.MemoServicePrice
		memo[core.BlockMemoKeyBlock] = strconv.FormatInt(currentBlock, 10)
		memo[core.BlockMemoKeyPrice] = ticker.Price.Truncate(8).String()
		memoStr, err := w.BlockService.FormatBlockMemo(ctx, memo)
		if err != nil {
			log.Errorln("new block memo error:", err)
			return err
		}

		transferInput.Memo = memoStr
		_, err = w.BlockWallet.Transfer(ctx, &transferInput, w.Config.BlockWallet.Pin)

		if err != nil {
			log.Errorln("transfer new block error:", err)
			return err
		}
	}

	return nil
}
