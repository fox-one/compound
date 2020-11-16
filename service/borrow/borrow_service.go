package borrow

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"errors"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

type borrowService struct {
	config         *core.Config
	mainWallet     *core.Wallet
	blockWallet    *core.Wallet
	marketStore    core.IMarketStore
	borrowStore    core.IBorrowStore
	blockService   core.IBlockService
	priceService   core.IPriceOracleService
	walletService  core.IWalletService
	accountService core.IAccountService
}

// New new borrow service
func New(cfg *core.Config,
	mainWallet *core.Wallet,
	blockWallet *core.Wallet,
	marketStore core.IMarketStore,
	borrowStore core.IBorrowStore,
	blockService core.IBlockService,
	priceService core.IPriceOracleService,
	walletService core.IWalletService,
	accountService core.IAccountService) core.IBorrowService {
	return &borrowService{
		config:         cfg,
		mainWallet:     mainWallet,
		blockWallet:    blockWallet,
		marketStore:    marketStore,
		borrowStore:    borrowStore,
		blockService:   blockService,
		priceService:   priceService,
		walletService:  walletService,
		accountService: accountService,
	}
}

func (s *borrowService) Repay(ctx context.Context, amount decimal.Decimal, borrow *core.Borrow) (string, error) {
	market, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
	if e != nil {
		return "", e
	}

	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceRepay
	action[core.ActionKeyBorrowTrace] = borrow.Trace

	memoStr, e := action.Format()
	if e != nil {
		return "", e
	}

	return s.walletService.PaySchemaURL(amount, market.AssetID, s.mainWallet.Client.ClientID, id.GenTraceID(), memoStr)
}

func (s *borrowService) Borrow(ctx context.Context, borrowAmount decimal.Decimal, userID string, market *core.Market) error {
	if !s.BorrowAllowed(ctx, borrowAmount, userID, market) {
		return errors.New("borrow not allowed")
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return e
	}

	trace := id.UUIDFromString(fmt.Sprintf("borrow-%s-%s-%d", userID, market.Symbol, curBlock))
	input := mixin.TransferInput{
		AssetID:    s.config.App.BlockAssetID,
		OpponentID: s.mainWallet.Client.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    trace,
	}

	if s.walletService.VerifyPayment(ctx, &input) {
		return errors.New("borrow exists")
	}

	memo := make(core.Action)
	memo[core.ActionKeyService] = core.ActionServiceBorrow
	memo[core.ActionKeyAmount] = borrowAmount.String()
	memo[core.ActionKeySymbol] = market.Symbol
	memo[core.ActionKeyUser] = userID

	memoStr, e := memo.Format()
	if e != nil {
		return e
	}

	input.Memo = memoStr

	_, e = s.blockWallet.Client.Transfer(ctx, &input, s.config.BlockWallet.Pin)
	if e != nil {
		return e
	}

	return nil
}

func (s *borrowService) BorrowAllowed(ctx context.Context, borrowAmount decimal.Decimal, userID string, market *core.Market) bool {
	log := logger.FromContext(ctx)

	if borrowAmount.LessThanOrEqual(decimal.Zero) {
		log.Errorln("invalid borrow amount")
		return false
	}

	// check borrow cap
	supplies := market.TotalCash.Sub(market.Reserves)
	if supplies.LessThan(market.BorrowCap) {
		log.Errorln("insufficient market cash")
		return false
	}

	if borrowAmount.GreaterThan(supplies.Sub(market.BorrowCap)) {
		log.Errorln("insufficient market cash")
		return false
	}

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		log.Errorln(e)
		return false
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		log.Errorln(e)
		return false
	}

	price, e := s.priceService.GetUnderlyingPrice(ctx, market.Symbol, curBlock)
	if e != nil {
		log.Errorln(e)
		return false
	}

	borrowValue := borrowAmount.Mul(price)
	if borrowValue.GreaterThan(liquidity) {
		log.Errorln("insufficient liquidity")
		return false
	}

	return true
}

func (s *borrowService) MaxBorrow(ctx context.Context, userID string, market *core.Market) (decimal.Decimal, error) {
	// check borrow cap
	supplies := market.TotalCash.Sub(market.Reserves)
	if supplies.LessThan(market.BorrowCap) {
		return decimal.Zero, errors.New("insufficient market cash")
	}

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return decimal.Zero, e
	}

	price, e := s.priceService.GetUnderlyingPrice(ctx, market.Symbol, curBlock)
	if e != nil {
		return decimal.Zero, e
	}

	borrowAmount := liquidity.Div(price)

	return borrowAmount, nil
}

func (s *borrowService) UpdateMarketInterestIndex(ctx context.Context, db *db.DB, market *core.Market, blockNum int64) error {
	borrows, e := s.borrowStore.FindBySymbol(ctx, market.Symbol)
	if e != nil {
		return e
	}

	for _, borrow := range borrows {
		deltaBlocks := blockNum - borrow.BlockNum
		if deltaBlocks > 0 && borrow.Principal.GreaterThan(decimal.Zero) {
			index := borrow.InterestIndex
			if index.LessThanOrEqual(decimal.Zero) {
				index = decimal.NewFromInt(1)
			}
			onePlusBlocksTimesRate := decimal.NewFromInt(1).Add(market.BorrowRatePerBlock.Mul(decimal.NewFromInt(deltaBlocks)))
			newIndex := index.Mul(onePlusBlocksTimesRate)
			borrow.InterestIndex = newIndex
			borrow.BlockNum = blockNum

			e = s.borrowStore.Update(ctx, db, borrow)
			if e != nil {
				return e
			}
		}
	}
	return nil
}
