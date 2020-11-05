package interest

import (
	"compound/core"
	"compound/pkg/concurrency"
	"compound/pkg/id"
	"compound/worker"
	"context"
	"fmt"
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
	Config        *core.Config
	MainWallet    *core.Wallet
	BlockWallet   *core.Wallet
	MarketStore   core.IMarketStore
	SupplyStore   core.ISupplyStore
	BorrowStore   core.IBorrowStore
	BlockService  core.IBlockService
	MarketService core.IMarketService
	WalletService core.IWalletService
}

// New new block worker
func New(cfg *core.Config,
	mainWallet *core.Wallet,
	blockWallet *core.Wallet,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	blockService core.IBlockService,
	marketService core.IMarketService,
	walletService core.IWalletService) *Worker {
	job := Worker{
		Config:        cfg,
		MainWallet:    mainWallet,
		BlockWallet:   blockWallet,
		MarketStore:   marketStore,
		SupplyStore:   supplyStore,
		BorrowStore:   borrowStore,
		BlockService:  blockService,
		MarketService: marketService,
		WalletService: walletService,
	}

	l, _ := time.LoadLocation(job.Config.App.Location)
	job.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 1s"
	job.Cron.AddFunc(spec, job.Run)
	job.OnWork = func() error {
		return job.onWork(context.Background())
	}

	return &job
}

func (w *Worker) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "interest")

	currentBlock, e := w.BlockService.CurrentBlock(ctx)
	if e != nil {
		log.Errorln(e)
		return e
	}

	markets, e := w.markets(ctx)
	if e != nil {
		return e
	}

	golimit := concurrency.DefaultGoLimit
	wg := sync.WaitGroup{}

	// supply interest
	supplies, e := w.SupplyStore.All(ctx)
	if e == nil {
		for _, supply := range supplies {
			wg.Add(1)
			go func(ctx context.Context, markets map[string]*core.Market, supply *core.Supply, currentBlock int64) {
				defer wg.Done()
				defer golimit.Done()

				w.calculateSupplyInterest(ctx, markets, supply, currentBlock)
			}(ctx, markets, supply, currentBlock)
		}
	}

	// borrow interest
	borrows, e := w.BorrowStore.All(ctx)
	if e == nil {
		for _, borrow := range borrows {
			wg.Add(1)
			go func(ctx context.Context, markets map[string]*core.Market, borrow *core.Borrow, currentBlock int64) {
				defer wg.Done()
				defer golimit.Done()
				w.calculateBorrowInterest(ctx, markets, borrow, currentBlock)
			}(ctx, markets, borrow, currentBlock)
		}
	}

	if supplies != nil {
		for i := range supplies {
			log.Infoln("supply index:", i)
			golimit.Add()
		}
	}

	if borrows != nil {
		for i := range borrows {
			log.Infoln("borrow index:", i)
			golimit.Add()
		}
	}

	golimit.Close()

	wg.Wait()

	return nil
}

func (w *Worker) calculateSupplyInterest(ctx context.Context, markets map[string]*core.Market, supply *core.Supply, currentBlock int64) {
	market, found := markets[supply.Symbol]
	if !found {
		return
	}

	rate, e := w.MarketService.CurSupplyRatePerBlock(ctx, market)
	if e != nil {
		return
	}

	interest := supply.Principal.Mul(rate)

	traceID := id.UUIDFromString(fmt.Sprintf("supply-interest-%s-%s-%d", supply.UserID, supply.Symbol, currentBlock))
	input := mixin.TransferInput{
		AssetID:    w.Config.App.BlockAssetID,
		OpponentID: w.MainWallet.Client.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    traceID,
	}

	if !w.WalletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceSupplyInterest
		action[core.ActionKeyUser] = supply.UserID
		action[core.ActionKeySymbol] = supply.Symbol
		action[core.ActionKeyAmount] = interest.Truncate(16).String()

		memoStr, e := action.Format()
		if e != nil {
			return
		}
		input.Memo = memoStr
		_, e = w.BlockWallet.Client.Transfer(ctx, &input, w.BlockWallet.Pin)
		if e != nil {
			return
		}
	}
}

func (w *Worker) calculateBorrowInterest(ctx context.Context, markets map[string]*core.Market, borrow *core.Borrow, currentBlock int64) {
	market, found := markets[borrow.Symbol]
	if !found {
		return
	}

	rate, e := w.MarketService.CurBorrowRatePerBlock(ctx, market)
	if e != nil {
		return
	}

	interest := borrow.Principal.Mul(rate)

	traceID := id.UUIDFromString(fmt.Sprintf("borrow-interest-%s-%s-%d", borrow.UserID, borrow.Symbol, currentBlock))
	input := mixin.TransferInput{
		AssetID:    w.Config.App.BlockAssetID,
		OpponentID: w.MainWallet.Client.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    traceID,
	}

	if !w.WalletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceBorrowInterest
		action[core.ActionKeyUser] = borrow.UserID
		action[core.ActionKeySymbol] = borrow.Symbol
		action[core.ActionKeyAmount] = interest.Truncate(16).String()

		memoStr, e := action.Format()
		if e != nil {
			return
		}
		input.Memo = memoStr
		_, e = w.BlockWallet.Client.Transfer(ctx, &input, w.BlockWallet.Pin)
		if e != nil {
			return
		}
	}
}

func (w *Worker) markets(ctx context.Context) (map[string]*core.Market, error) {
	markets, e := w.MarketStore.All(ctx)
	if e != nil {
		return nil, e
	}

	maps := make(map[string]*core.Market)

	for _, m := range markets {
		maps[m.Symbol] = m
	}

	return maps, nil
}
