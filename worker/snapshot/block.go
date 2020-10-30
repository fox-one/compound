package snapshot

import (
	"compound/core"
	"context"
	"strconv"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

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

	block, err := strconv.ParseInt(blockMemo[core.ActionKeyBlock], 10, 64)
	if err != nil {
		return nil
	}

	service := blockMemo[core.ActionKeyService]
	if service == core.ActionServicePrice {
		// cache price

		symbol := blockMemo[core.ActionKeySymbol]
		price, err := decimal.NewFromString(blockMemo[core.ActionKeyPrice])
		if err != nil {
			return nil
		}

		w.priceService.Save(ctx, symbol, price, block)
	} else if service == core.ActionServiceMarket {
		symbol := blockMemo[core.ActionKeySymbol]

		//utilization rate
		utilizationRate, err := decimal.NewFromString(blockMemo[core.ActionKeyUtilizationRate])
		if err != nil {
			return nil
		}

		w.marketService.SaveUtilizationRate(ctx, symbol, utilizationRate, block)

		// borrow rate
		borrowRate, err := decimal.NewFromString(blockMemo[core.ActionKeyBorrowRate])
		if err != nil {
			return nil
		}

		w.marketService.SaveBorrowRatePerBlock(ctx, symbol, borrowRate, block)

		// supply rate
		supplyRate, err := decimal.NewFromString(blockMemo[core.ActionKeySupplyRate])
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
