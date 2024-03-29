package payee

import (
	"compound/core"
	"compound/pkg/compound"
	"context"

	"github.com/fox-one/pkg/logger"
)

func (w *Payee) mustGetMarket(ctx context.Context, asset string) (*core.Market, error) {
	log := logger.FromContext(ctx)

	market, err := w.marketStore.Find(ctx, asset)
	if err != nil {
		log.WithError(err).Errorln("markets.Find")
		return nil, err
	}

	if err := compound.Require(market.ID > 0, "payee/market-not-found"); err != nil {
		log.WithError(err).Infoln("skip")
		return nil, err
	}

	return market, nil
}

func (w *Payee) mustGetMarketWithCToken(ctx context.Context, ctoken string) (*core.Market, error) {
	log := logger.FromContext(ctx)

	market, err := w.marketStore.FindByCToken(ctx, ctoken)
	if err != nil {
		log.WithError(err).Errorln("markets.FindByCToken")
		return nil, err
	}

	if err := compound.Require(market.ID > 0, "payee/market-not-found"); err != nil {
		log.WithError(err).Infoln("skip")
		return nil, err
	}

	return market, nil
}

func (w *Payee) mustGetSupply(ctx context.Context, user, asset string) (*core.Supply, error) {
	log := logger.FromContext(ctx)

	supply, err := w.supplyStore.Find(ctx, user, asset)
	if err != nil {
		log.WithError(err).Errorln("supplies.Find")
		return nil, err
	}

	if w.sysversion < 5 {
		if err := compound.Require(supply.ID > 0, "payee/supply-not-found"); err != nil {
			log.WithError(err).Infoln("skip")
			return nil, err
		}
	} else {
		if err := compound.Require(supply.ID > 0, "payee/supply-not-found", compound.FlagRefund); err != nil {
			log.WithError(err).Infoln("skip")
			return nil, err
		}
	}

	return supply, nil
}

func (w *Payee) getOrCreateSupply(ctx context.Context, user, asset string) (*core.Supply, error) {
	log := logger.FromContext(ctx)

	supply, err := w.supplyStore.Find(ctx, user, asset)
	if err != nil {
		log.WithError(err).Errorln("supplies.Find")
		return nil, err
	}

	if supply.ID == 0 {
		supply = &core.Supply{
			UserID:        user,
			CTokenAssetID: asset,
		}
		if err := w.supplyStore.Create(ctx, supply); err != nil {
			log.WithError(err).Errorln("supplies.Create")
			return nil, err
		}
	}
	return supply, nil
}

func (w *Payee) mustGetBorrow(ctx context.Context, user, asset string) (*core.Borrow, error) {
	log := logger.FromContext(ctx)

	borrow, err := w.borrowStore.Find(ctx, user, asset)
	if err != nil {
		log.WithError(err).Errorln("borrows.Find")
		return nil, err
	}

	if w.sysversion < 5 {
		if err := compound.Require(borrow.ID > 0, "payee/borrow-not-found"); err != nil {
			log.WithError(err).Infoln("skip")
			return nil, err
		}
	} else {
		if err := compound.Require(borrow.ID > 0, "payee/borrow-not-found", compound.FlagRefund); err != nil {
			log.WithError(err).Infoln("skip")
			return nil, err
		}
	}

	return borrow, nil
}

func (w *Payee) getOrCreateBorrow(ctx context.Context, user, asset string) (*core.Borrow, error) {
	log := logger.FromContext(ctx)

	borrow, err := w.borrowStore.Find(ctx, user, asset)
	if err != nil {
		log.WithError(err).Errorln("borrows.Find")
		return nil, err
	}

	if borrow.ID == 0 {
		borrow = &core.Borrow{
			UserID:  user,
			AssetID: asset,
		}
		if err := w.borrowStore.Create(ctx, borrow); err != nil {
			log.WithError(err).Errorln("borrows.Create")
			return nil, err
		}
	}

	return borrow, nil
}

func (w *Payee) mustGetProposal(ctx context.Context, trace string) (*core.Proposal, error) {
	log := logger.FromContext(ctx)

	proposal, err := w.proposalStore.Find(ctx, trace)
	if err != nil {
		log.WithError(err).Errorln("proposals.Find")
		return nil, err
	}

	if err := compound.Require(proposal.ID > 0, "payee/proposal-not-found"); err != nil {
		log.WithError(err).Infoln("skip")
		return nil, err
	}

	return proposal, nil
}
