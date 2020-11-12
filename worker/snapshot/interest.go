package snapshot

import (
	"compound/core"
	"context"
	"strconv"

	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

var handleBorrowInterestEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	userID := action[core.ActionKeyUser]
	symbol := action[core.ActionKeySymbol]
	blockNum, e := strconv.ParseInt(action[core.ActionKeyBlock], 10, 64)
	if e != nil {
		return e
	}
	interest, e := decimal.NewFromString(action[core.ActionKeyAmount])
	if e != nil {
		return e
	}

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, symbol)
	if e != nil {
		return e
	}

	return w.db.Tx(func(tx *db.DB) error {
		market.BlockNumber = blockNum
		market.Reserves = market.Reserves.Add(interest.Mul(market.ReserveFactor))
		e = w.marketStore.Update(ctx, tx, market)
		if e != nil {
			return e
		}

		borrow.Principal = borrow.Principal.Add(interest)
		borrow.InterestAccumulated = borrow.InterestAccumulated.Add(interest)
		e = w.borrowStore.Update(ctx, tx, borrow)
		if e != nil {
			return e
		}

		return nil
	})
}
