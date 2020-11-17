package borrow

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
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
	marketService  core.IMarketService
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
	accountService core.IAccountService,
	marketService core.IMarketService) core.IBorrowService {
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
		marketService:  marketService,
	}
}

func (s *borrowService) Repay(ctx context.Context, amount decimal.Decimal, borrow *core.Borrow) (string, error) {
	market, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
	if e != nil {
		return "", e
	}

	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceRepay

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

	blockNum, e := s.blockService.GetBlock(ctx, time.Now())
	if e != nil {
		return e
	}

	trace := id.UUIDFromString(fmt.Sprintf("borrow-%s-%s-%d", userID, market.Symbol, blockNum))
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

	blockNum, e := s.blockService.GetBlock(ctx, time.Now())
	if e != nil {
		log.Errorln(e)
		return false
	}

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID, blockNum)
	if e != nil {
		log.Errorln(e)
		return false
	}

	price, e := s.priceService.GetCurrentUnderlyingPrice(ctx, market)
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

	blockNum, e := s.blockService.GetBlock(ctx, time.Now())
	if e != nil {
		return decimal.Zero, e
	}

	// check liquidity
	liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID, blockNum)
	if e != nil {
		return decimal.Zero, e
	}

	price, e := s.priceService.GetCurrentUnderlyingPrice(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	borrowAmount := liquidity.Div(price)

	return borrowAmount, nil
}

func (s *borrowService) BorrowBalance(ctx context.Context, borrow *core.Borrow, market *core.Market) (decimal.Decimal, error) {
	market, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
	if e != nil {
		return decimal.Zero, e
	}

	principalTimesIndex := borrow.Principal.Mul(market.BorrowIndex)
	result := principalTimesIndex.Div(borrow.InterestIndex)

	return result, nil
}
