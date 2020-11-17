package market

import (
	"compound/core"
	"compound/internal/compound"
	"fmt"
	"time"

	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/go-redis/redis"
	"github.com/shopspring/decimal"
)

type service struct {
	Redis       *redis.Client
	mainWallet  *core.Wallet
	marketStore core.IMarketStore
	borrowStore core.IBorrowStore
	blockSrv    core.IBlockService
	priceSrv    core.IPriceOracleService
}

// New new market service
func New(
	redis *redis.Client,
	mainWallet *core.Wallet,
	marketStr core.IMarketStore,
	borrowStore core.IBorrowStore,
	blockSrv core.IBlockService,
	priceSrv core.IPriceOracleService,
) core.IMarketService {
	return &service{
		Redis:       redis,
		mainWallet:  mainWallet,
		marketStore: marketStr,
		borrowStore: borrowStore,
		blockSrv:    blockSrv,
		priceSrv:    priceSrv,
	}
}

//资金使用率，同一个block里保持一致，该数据会影响到借款和存款利率的计算
func (s *service) curUtilizationRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	rate := compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves)
	return rate, nil
}

func (s *service) curExchangeRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	if market.CTokens.LessThanOrEqual(decimal.Zero) {
		return market.InitExchangeRate, nil
	}

	rate := compound.GetExchangeRate(market.TotalCash, market.TotalBorrows, market.Reserves, market.CTokens, market.InitExchangeRate)

	return rate, nil
}

// 借款年利率
func (s *service) CurBorrowRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	borrowRatePerBlock, e := s.curBorrowRatePerBlock(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return borrowRatePerBlock.Mul(compound.BlocksPerYear), nil
}

// 借款块利率, 同一个block里保持一致
func (s *service) curBorrowRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.curUtilizationRate(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetBorrowRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink)

	return rate, nil
}

// 存款年利率
func (s *service) CurSupplyRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	supplyRatePerBlock, e := s.curSupplyRatePerBlock(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return supplyRatePerBlock.Mul(compound.BlocksPerYear), nil
}

// 存款块利率, 同一个block里保持一致
func (s *service) curSupplyRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.curUtilizationRate(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetSupplyRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink, market.ReserveFactor)

	return rate, nil
}

// 总借出量
func (s *service) CurTotalBorrow(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.TotalBorrows, nil
}

// 总保留金
func (s *service) CurTotalReserves(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.Reserves, nil
}

func (s *service) KeppFlywheelMoving(ctx context.Context, db *db.DB, market *core.Market, time time.Time) error {
	blockNum, e := s.blockSrv.GetBlock(ctx, time)
	if e != nil {
		return e
	}
	//utilization rate
	uRate, e := s.curUtilizationRate(ctx, market)
	if e != nil {
		return e
	}

	//exchange rate
	exchangeRate, e := s.curExchangeRate(ctx, market)
	if e != nil {
		return e
	}

	supplyRate, e := s.curSupplyRatePerBlock(ctx, market)
	if e != nil {
		return e
	}

	borrowRate, e := s.curBorrowRatePerBlock(ctx, market)
	if e != nil {
		return e
	}

	market.BlockNumber = blockNum
	market.UtilizationRate = uRate
	market.ExchangeRate = exchangeRate
	market.SupplyRatePerBlock = supplyRate
	market.BorrowRatePerBlock = borrowRate

	e = s.marketStore.Update(ctx, db, market)
	if e != nil {
		return e
	}

	return nil
}

func (s *service) borrowRateCacheKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:brate:%s:%d", symbol, block)
}

func (s *service) supplyRateCacheKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:srate:%s:%d", symbol, block)
}

func (s *service) utilizationRateCacheKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:urate:%s:%d", symbol, block)
}

func (s *service) exchangeRateCacheKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:erate:%s:%d", symbol, block)
}
