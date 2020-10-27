package block

import (
	"compound/core"
	"compound/pkg/id"
	"compound/worker"
	"context"
	"fmt"
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
	Dapp         *mixin.Client
	BlockWallet  *mixin.Client
	BlockService core.IBlockService
}

// New new block worker
func New(cfg *core.Config, dapp *mixin.Client, blockWallet *mixin.Client, blockService core.IBlockService) *Worker {
	job := Worker{
		Config:       cfg,
		Dapp:         dapp,
		BlockWallet:  blockWallet,
		BlockService: blockService,
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
	log := logger.FromContext(ctx).WithField("worker", "block")

	currentBlock, err := w.BlockService.CurrentBlock(ctx)
	// 检查当前区块是否已创建
	str := fmt.Sprintf("foxone-compound-block-%d", currentBlock)
	traceID := id.UUIDFromString(str)
	transferInput := mixin.TransferInput{
		AssetID:    w.Config.App.BlockAssetID,
		OpponentID: w.Config.Mixin.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    traceID,
	}
	payment, err := w.Dapp.VerifyPayment(ctx, transferInput)
	if err != nil {
		log.Errorln("verifypayment error:", err)
		return err
	}

	if payment.Status == "paid" {
		log.Infoln("block exists")
	} else {
		//create new block
		memoStr, err := w.BlockService.NewBlockMemo(ctx, currentBlock)
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
