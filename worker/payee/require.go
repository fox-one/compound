package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"

	"github.com/fox-one/pkg/logger"
)

func (w *Payee) requireMarket(ctx context.Context, asset string) (*core.Market, error) {
	log := logger.FromContext(ctx)

	market, err := w.marketStore.Find(ctx, asset)
	if err != nil {
		log.WithError(err).Errorln("markets.Find")
		return nil, err
	}

	if err := compound.Require(market.ID > 0, "payee/skip/market-not-found", compound.FlagNoisy); err != nil {
		log.WithError(err).Infoln("skip")
		return nil, err
	}

	return market, nil
}

func (w *Payee) requireSupply(ctx context.Context, user, asset string) (*core.Supply, error) {
	log := logger.FromContext(ctx)

	supply, err := w.supplyStore.Find(ctx, user, asset)
	if err != nil {
		log.WithError(err).Errorln("supplies.Find")
		return nil, err
	}

	if err := compound.Require(supply.ID > 0, "payee/skip/supply-not-found", compound.FlagNoisy); err != nil {
		log.WithError(err).Infoln("skip")
		return nil, err
	}

	return supply, nil
}

func (w *Payee) requireBorrow(ctx context.Context, user, asset string) (*core.Borrow, error) {
	log := logger.FromContext(ctx)

	borrow, err := w.borrowStore.Find(ctx, user, asset)
	if err != nil {
		log.WithError(err).Errorln("borrows.Find")
		return nil, err
	}

	if err := compound.Require(borrow.ID > 0, "payee/skip/borrow-not-found", compound.FlagNoisy); err != nil {
		log.WithError(err).Infoln("skip")
		return nil, err
	}

	return borrow, nil
}
