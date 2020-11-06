package snapshot

import (
	"compound/core"
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

var handleSupplyInterestEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := action[core.ActionKeyUser]
	symbol := action[core.ActionKeySymbol]
	interest, e := decimal.NewFromString(action[core.ActionKeyAmount])
	if e != nil {
		return e
	}

	market, e := w.marketStore.Find(ctx, "", symbol)
	if e != nil {
		return e
	}

	supply, e := w.supplyStore.Find(ctx, userID, symbol)
	if e != nil {
		return e
	}

	return w.db.Tx(func(tx *db.DB) error {
		market.TotalSupplyInterest = market.TotalSupplyInterest.Add(interest)
		e = w.marketStore.Update(ctx, tx, market)
		if e != nil {
			return e
		}

		supply.Principal = supply.Principal.Add(interest)
		supply.InterestAccumulated = supply.InterestAccumulated.Add(interest)
		supply.InterestBalance = supply.InterestBalance.Add(interest)
		e = w.supplyStore.Update(ctx, tx, supply)
		if e != nil {
			return e
		}

		return nil
	})
}

var handleBorrowInterestEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := action[core.ActionKeyUser]
	symbol := action[core.ActionKeySymbol]
	interest, e := decimal.NewFromString(action[core.ActionKeyAmount])
	if e != nil {
		return e
	}

	market, e := w.marketStore.Find(ctx, "", symbol)
	if e != nil {
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, symbol)
	if e != nil {
		return e
	}

	return w.db.Tx(func(tx *db.DB) error {
		market.TotalBorrowInterest = market.TotalBorrowInterest.Add(interest)
		e = w.marketStore.Update(ctx, tx, market)
		if e != nil {
			return e
		}

		borrow.Principal = borrow.Principal.Add(interest)
		borrow.InterestAccumulated = borrow.InterestAccumulated.Add(interest)
		borrow.InterestBalance = borrow.InterestBalance.Add(interest)
		e = w.borrowStore.Update(ctx, tx, borrow)
		if e != nil {
			return e
		}

		return nil
	})
}
