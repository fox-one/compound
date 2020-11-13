package block

import (
	"compound/core"
	"compound/pkg/id"
	"compound/worker"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

//Worker block worker
type Worker struct {
	worker.BaseJob
	Config       *core.Config
	MainWallet   *core.Wallet
	BlockWallet  *core.Wallet
	BlockService core.IBlockService
}

// New new block worker
func New(cfg *core.Config, mainWallet *core.Wallet, blockWallet *core.Wallet, blockService core.IBlockService) *Worker {
	job := Worker{
		Config:       cfg,
		MainWallet:   mainWallet,
		BlockWallet:  blockWallet,
		BlockService: blockService,
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
	log := logger.FromContext(ctx).WithField("worker", "block")

	currentBlock, err := w.BlockService.CurrentBlock(ctx)
	if err != nil {
		log.Errorln(err)
		return err
	}
	// 检查当前区块是否已创建
	str := fmt.Sprintf("foxone-compound-block-%d", currentBlock)
	traceID := id.UUIDFromString(str)
	transferInput := mixin.TransferInput{
		AssetID:    w.Config.App.BlockAssetID,
		OpponentID: w.MainWallet.Client.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    traceID,
	}
	payment, err := w.MainWallet.Client.VerifyPayment(ctx, transferInput)
	if err != nil {
		log.Errorln("verifypayment error:", err)
		return err
	}

	if payment.Status == "paid" {
		log.Infoln("block exists")
	} else {
		//create new block
		memo := make(core.Action)
		memo[core.ActionKeyService] = core.ActionServiceBlock
		memo[core.ActionKeyBlock] = strconv.FormatInt(currentBlock, 10)
		memoStr, err := memo.Format()
		if err != nil {
			log.Errorln("new block memo error:", err)
			return err
		}

		transferInput.Memo = memoStr
		_, err = w.BlockWallet.Client.Transfer(ctx, &transferInput, w.Config.BlockWallet.Pin)

		if err != nil {
			log.Errorln("transfer new block error:", err)
			return err
		}
	}

	return nil
}
