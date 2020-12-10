package priceoracle

import (
	"compound/core"
	"compound/worker"
	"context"
	"sync"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

//Worker block worker
type Worker struct {
	worker.BaseJob
	system             *core.System
	MainWallet         *core.Wallet
	BlockWallet        *core.Wallet
	Config             *core.Config
	MarketStore        core.IMarketStore
	BlockService       core.IBlockService
	PriceOracleService core.IPriceOracleService
}

// New new block worker
func New(system *core.System, mainWallet *core.Wallet, blockWallet *core.Wallet, cfg *core.Config, marketStore core.IMarketStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService) *Worker {
	job := Worker{
		MainWallet:         mainWallet,
		BlockWallet:        blockWallet,
		Config:             cfg,
		MarketStore:        marketStore,
		BlockService:       blockSrv,
		PriceOracleService: priceSrv,
	}

	l, _ := time.LoadLocation(job.Config.Location)
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
	// log := logger.FromContext(ctx).WithField("worker", "priceoracle")

	// blockNum, err := w.BlockService.GetBlock(ctx, time.Now())
	// if err != nil {
	// 	log.Errorln(err)
	// 	return err
	// }

	// str := fmt.Sprintf("foxone-compound-price-%s-%d", market.Symbol, blockNum)
	// traceID := id.UUIDFromString(str)
	// transferInput := mixin.TransferInput{
	// 	AssetID:    w.Config.Group.Vote.Asset,
	// 	OpponentID: w.MainWallet.Client.ClientID,
	// 	Amount:     w.system.VoteAmount,
	// 	TraceID:    traceID,
	// }

	// //create new block
	// memo := make(core.Action)
	// memo[core.ActionKeyService] = core.ActionServicePrice
	// memo[core.ActionKeyBlock] = strconv.FormatInt(blockNum, 10)
	// memo[core.ActionKeySymbol] = market.Symbol
	// memo[core.ActionKeyPrice] = ticker.Price.Truncate(8).String()
	// memoStr, err := memo.Format()
	// if err != nil {
	// 	log.Errorln("new block memo error:", err)
	// 	return err
	// }

	// transferInput.Memo = memoStr
	// if _, err = w.BlockWallet.Client.Transfer(ctx, &transferInput, w.BlockWallet.Pin); err != nil {
	// 	log.Errorln("transfer new block error:", err)
	// 	return err
	// }

	return nil
}
