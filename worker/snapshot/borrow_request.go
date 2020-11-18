package snapshot

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"fmt"
	"strings"

	"github.com/fox-one/mixin-sdk-go"
)

var handleRequestBorrowEvent = func(ctx context.Context, w *Worker, action core.Action, snapshot *core.Snapshot) error {
	if snapshot.AssetID != w.config.App.GasAssetID {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrOperationForbidden)
	}

	userID := action[core.ActionKeyUser]
	symbol := strings.ToUpper(action[core.ActionKeySymbol])

	market, e := w.marketStore.FindBySymbol(ctx, symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrMarketNotFound)
	}

	// accrue interest
	if e = w.marketService.AccrueInterest(ctx, w.db, market, snapshot.CreatedAt); e != nil {
		return e
	}

	borrow, e := w.borrowStore.Find(ctx, userID, symbol)
	if e != nil {
		return handleRefundEvent(ctx, w, action, snapshot, core.ErrBorrowNotFound)
	}

	borrowBalance, e := w.borrowService.BorrowBalance(ctx, borrow, market)
	if e != nil {
		return e
	}

	trace := id.UUIDFromString(fmt.Sprintf("borrow-%s", snapshot.TraceID))
	input := mixin.TransferInput{
		AssetID:    w.config.App.GasAssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     core.GasCost,
		TraceID:    trace,
	}

	if !w.walletService.VerifyPayment(ctx, &input) {
		action := core.NewAction()
		action[core.ActionKeyService] = core.ActionServiceBorrowResponse
		action[core.ActionKeyUser] = userID
		action[core.ActionKeySymbol] = symbol
		action[core.ActionKeyAmount] = borrow.Principal.String()
		action[core.ActionKeyInterestIndex] = borrow.InterestIndex.String()
		action[core.ActionKeyBorrowBalance] = borrowBalance.Truncate(8).String()
		memoStr, e := action.Format()
		if e != nil {
			return e
		}
		input.Memo = memoStr
		if _, e = w.mainWallet.Client.Transfer(ctx, &input, w.mainWallet.Pin); e != nil {
			return e
		}
	}

	return nil
}
