package priceoracle

import (
	"compound/core"
	"compound/core/proposal"
	"compound/pkg/id"
	"compound/pkg/mtg"
	"compound/worker"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

//Worker block worker
type Worker struct {
	worker.TickWorker
	System             *core.System
	Dapp               *core.Wallet
	MarketStore        core.IMarketStore
	PriceStore         core.IPriceStore
	BlockService       core.IBlockService
	PriceOracleService core.IPriceOracleService
}

// New new block worker
func New(system *core.System, dapp *core.Wallet, marketStore core.IMarketStore, priceStr core.IPriceStore, blockSrv core.IBlockService, priceSrv core.IPriceOracleService) *Worker {
	job := Worker{
		System:             system,
		Dapp:               dapp,
		MarketStore:        marketStore,
		PriceStore:         priceStr,
		BlockService:       blockSrv,
		PriceOracleService: priceSrv,
	}

	return &job
}

// Run run worker
func (w *Worker) Run(ctx context.Context) error {
	return w.StartTick(ctx, func(ctx context.Context) error {
		return w.onWork(ctx)
	})
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
			if !w.isPriceProvided(ctx, market) {
				// do real work
				ticker, e := w.PriceOracleService.PullPriceTicker(ctx, market.AssetID, time.Now())
				if e != nil {
					log.Errorln("pull price ticker error:", e)
					return
				}
				if ticker.Price.LessThanOrEqual(decimal.Zero) {
					log.Errorln("invalid ticker price:", ticker.Symbol, ":", ticker.Price)
					return
				}

				w.pushPriceOnChain(ctx, market, ticker)
			}
		}(m)
	}

	wg.Wait()

	return nil
}

func (w *Worker) isPriceProvided(ctx context.Context, market *core.Market) bool {
	log := logger.FromContext(ctx).WithField("worker", "priceoracle")

	curBlockNum, e := w.BlockService.GetBlock(ctx, time.Now())
	if e != nil {
		log.WithError(e).Errorln("GetBlock err")
		return false
	}

	price, _, e := w.PriceStore.FindByAssetBlock(ctx, market.AssetID, curBlockNum)
	if e != nil {
		log.WithError(e).Errorln("findByAssetBlock err")
		return false
	}

	var priceTickers []*core.PriceTicker
	if e = json.Unmarshal(price.Content, &priceTickers); e != nil {
		return false
	}

	for _, p := range priceTickers {
		if p.Provider == w.System.ClientID {
			return true
		}
	}

	return false
}

func (w *Worker) pushPriceOnChain(ctx context.Context, market *core.Market, ticker *core.PriceTicker) error {
	log := logger.FromContext(ctx).WithField("worker", "priceoracle")

	blockNum, e := w.BlockService.GetBlock(ctx, time.Now())
	if e != nil {
		log.Errorln(e)
		return e
	}

	traceID := id.UUIDFromString(fmt.Sprintf("price-%s-%s-%d", w.System.ClientID, market.AssetID, blockNum))
	// transfer
	providePriceReq := proposal.ProvidePriceReq{
		AssetID: market.AssetID,
		Price:   ticker.Price,
	}

	memo, e := mtg.Encode(w.System.ClientID, traceID, int(core.ActionTypeProposalProvidePrice), providePriceReq)
	if e != nil {
		return e
	}
	sign := mtg.Sign(memo, w.System.SignKey)
	memo = mtg.Pack(memo, sign)

	input := mixin.TransferInput{
		AssetID: w.System.VoteAsset,
		Amount:  w.System.VoteAmount,
		TraceID: traceID,
		Memo:    base64.StdEncoding.EncodeToString(memo),
	}

	input.OpponentMultisig.Receivers = w.System.MemberIDs()
	input.OpponentMultisig.Threshold = w.System.Threshold

	// multisig transfer
	_, e = w.Dapp.Client.Transaction(ctx, &input, w.Dapp.Pin)
	if e != nil {
		return e
	}

	return nil
}
