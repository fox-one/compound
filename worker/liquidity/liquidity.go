package liquidity

import (
	"compound/core"
	"compound/pkg/concurrency"
	"compound/pkg/id"
	"compound/worker"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

//Worker block worker
type Worker struct {
	worker.BaseJob
	Config         *core.Config
	MainWallet     *core.Wallet
	BlockWallet    *core.Wallet
	MarketStore    core.IMarketStore
	SupplyStore    core.ISupplyStore
	BorrowStore    core.IBorrowStore
	BlockService   core.IBlockService
	MarketService  core.IMarketService
	WalletService  core.IWalletService
	AccountService core.IAccountService
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
	walletService core.IWalletService,
	accountService core.IAccountService) *Worker {
	job := Worker{
		Config:         cfg,
		MainWallet:     mainWallet,
		BlockWallet:    blockWallet,
		MarketStore:    marketStore,
		SupplyStore:    supplyStore,
		BorrowStore:    borrowStore,
		BlockService:   blockService,
		MarketService:  marketService,
		WalletService:  walletService,
		AccountService: accountService,
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
	curBlock, e := w.BlockService.CurrentBlock(ctx)
	if e != nil {
		return e
	}

	borrows, e := w.BorrowStore.All(ctx)
	if e != nil {
		return e
	}

	golimit := concurrency.DefaultGoLimit
	wg := sync.WaitGroup{}

	for _, borrow := range borrows {
		wg.Add(1)
		golimit.Add()
		go func(ctx context.Context, userID string, curBlock int64) {
			defer wg.Done()
			defer golimit.Done()
			w.calculateLiquidity(ctx, userID, curBlock)
		}(ctx, borrow.UserID, curBlock)
	}

	golimit.Close()
	wg.Wait()

	return nil
}

func (w *Worker) calculateLiquidity(ctx context.Context, userID string, curBlock int64) error {
	trace := id.UUIDFromString(fmt.Sprintf("liquidity-%s-%d", userID, curBlock))
	input := mixin.TransferInput{
		AssetID:    w.Config.App.BlockAssetID,
		OpponentID: w.MainWallet.Client.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    trace,
	}

	if !w.WalletService.VerifyPayment(ctx, &input) {
		liquidity, e := w.AccountService.CalculateAccountLiquidity(ctx, userID)
		if e != nil {
			return e
		}
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceLiquidity
		memo[core.ActionKeyBlock] = strconv.FormatInt(curBlock, 10)
		memo[core.ActionKeyAmount] = liquidity.Truncate(8).String()
		memo[core.ActionKeyUser] = userID
		memoStr, e := memo.Format()
		if e != nil {
			return e
		}

		input.Memo = memoStr
		_, e = w.BlockWallet.Client.Transfer(ctx, &input, w.BlockWallet.Pin)
		if e != nil {
			return e
		}
	}
	return nil
}
