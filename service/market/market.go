package market

import (
	"compound/core"
	"compound/internal/compound"
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

func (s *service) CurUtilizationRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	ur := market.UtilizationRate
	if ur.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}

	return ur, nil
}
func (s *service) CurExchangeRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	er := market.ExchangeRate
	if er.LessThanOrEqual(decimal.Zero) {
		return market.InitExchangeRate, nil
	}

	return er, nil
}
func (s *service) CurBorrowRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	br := market.BorrowRatePerBlock
	if br.LessThanOrEqual(decimal.Zero) {
		return s.curBorrowRatePerBlockInternal(ctx, market)
	}

	return br, nil
}
func (s *service) CurSupplyRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	sr := market.SupplyRatePerBlock
	if sr.LessThanOrEqual(decimal.Zero) {
		return s.curSupplyRatePerBlockInternal(ctx, market)
	}

	return sr, nil
}

//资金使用率，同一个block里保持一致，该数据会影响到借款和存款利率的计算
func (s *service) curUtilizationRateInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	rate := compound.UtilizationRate(market.TotalCash, market.TotalBorrows, market.Reserves)
	return rate, nil
}

func (s *service) curExchangeRateInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	if market.CTokens.LessThanOrEqual(decimal.Zero) {
		return market.InitExchangeRate, nil
	}

	rate := compound.GetExchangeRate(market.TotalCash, market.TotalBorrows, market.Reserves, market.CTokens, market.InitExchangeRate)

	return rate, nil
}

// 借款年利率
func (s *service) CurBorrowRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	borrowRatePerBlock, e := s.curBorrowRatePerBlockInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return borrowRatePerBlock.Mul(compound.BlocksPerYear), nil
}

// 借款块利率, 同一个block里保持一致
func (s *service) curBorrowRatePerBlockInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.curUtilizationRateInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetBorrowRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink)

	return rate, nil
}

// 存款年利率
func (s *service) CurSupplyRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	supplyRatePerBlock, e := s.curSupplyRatePerBlockInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return supplyRatePerBlock.Mul(compound.BlocksPerYear), nil
}

// 存款块利率, 同一个block里保持一致
func (s *service) curSupplyRatePerBlockInternal(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.curUtilizationRateInternal(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetSupplyRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink, market.ReserveFactor)

	return rate, nil
}

// 总借出量
func (s *service) CurTotalBorrows(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
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
	uRate, e := s.curUtilizationRateInternal(ctx, market)
	if e != nil {
		return e
	}

	//exchange rate
	exchangeRate, e := s.curExchangeRateInternal(ctx, market)
	if e != nil {
		return e
	}

	supplyRate, e := s.curSupplyRatePerBlockInternal(ctx, market)
	if e != nil {
		return e
	}

	borrowRate, e := s.curBorrowRatePerBlockInternal(ctx, market)
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

func (s *service) AccrueInterest(ctx context.Context, db *db.DB, market *core.Market, time time.Time) error {
	blockNumberPrior := market.BlockNumber

	blockNum, e := s.blockSrv.GetBlock(ctx, time)
	if e != nil {
		return e
	}

	blockDelta := blockNum - blockNumberPrior
	if blockDelta > 0 {
		borrowRate, e := s.curBorrowRatePerBlockInternal(ctx, market)
		if e != nil {
			return e
		}

		timesBorrowRate := borrowRate.Mul(decimal.NewFromInt(blockDelta))
		interestAccumulated := market.TotalBorrows.Mul(timesBorrowRate)
		totalBorrowsNew := interestAccumulated.Add(market.TotalBorrows)
		totalReservesNew := interestAccumulated.Mul(market.ReserveFactor).Add(market.Reserves)
		borrowIndexNew := market.BorrowIndex.Add(timesBorrowRate.Mul(market.BorrowIndex))

		market.BlockNumber = blockNum
		market.TotalBorrows = totalBorrowsNew
		market.Reserves = totalReservesNew
		market.BorrowIndex = borrowIndexNew
	}

	//utilization rate
	uRate, e := s.curUtilizationRateInternal(ctx, market)
	if e != nil {
		return e
	}

	//exchange rate
	exchangeRate, e := s.curExchangeRateInternal(ctx, market)
	if e != nil {
		return e
	}

	supplyRate, e := s.curSupplyRatePerBlockInternal(ctx, market)
	if e != nil {
		return e
	}

	borrowRate, e := s.curBorrowRatePerBlockInternal(ctx, market)
	if e != nil {
		return e
	}

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
