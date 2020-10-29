package market

import (
	"compound/core"
	"compound/internal/compound"
	"fmt"
	"time"

	"context"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis"
	"github.com/shopspring/decimal"
)

type service struct {
	Redis         *redis.Client
	mainWallet    *mixin.Client
	reserveWallet *mixin.Client
	marketstore   core.IMarketStore
	blockSrv      core.IBlockService
	priceSrv      core.IPriceOracleService
}

// New new market service
func New(
	redis *redis.Client,
	mainWallet *mixin.Client,
	reserveWallet *mixin.Client,
	marketStr core.IMarketStore,
	blockSrv core.IBlockService,
	priceSrv core.IPriceOracleService,
) core.IMarketService {
	return &service{
		Redis:         redis,
		mainWallet:    mainWallet,
		reserveWallet: reserveWallet,
		marketstore:   marketStr,
		blockSrv:      blockSrv,
		priceSrv:      priceSrv,
	}
}

func (s *service) SaveUtilizationRate(ctx context.Context, symbol string, rate decimal.Decimal, block int64) error {
	k := s.utilizationRateCacheKey(symbol, block)

	// not exists, add new
	if s.Redis.Exists(k).Val() == 0 {
		//new
		s.Redis.Set(k, []byte(rate.String()), time.Hour)
	}
	return nil
}
func (s *service) GetUtilizationRate(ctx context.Context, symbol string, block int64) (decimal.Decimal, error) {
	k := s.utilizationRateCacheKey(symbol, block)

	bs, e := s.Redis.Get(k).Bytes()
	if e != nil {
		return decimal.Zero, e
	}

	price, e := decimal.NewFromString(string(bs))
	if e != nil {
		return decimal.Zero, e
	}

	return price, nil
}

func (s *service) SaveBorrowRatePerBlock(ctx context.Context, symbol string, rate decimal.Decimal, block int64) error {
	k := s.borrowRateCacheKey(symbol, block)

	// not exists, add new
	if s.Redis.Exists(k).Val() == 0 {
		//new
		s.Redis.Set(k, []byte(rate.String()), time.Hour)
	}
	return nil
}
func (s *service) GetBorrowRatePerBlock(ctx context.Context, symbol string, block int64) (decimal.Decimal, error) {
	k := s.borrowRateCacheKey(symbol, block)

	bs, e := s.Redis.Get(k).Bytes()
	if e != nil {
		return decimal.Zero, e
	}

	price, e := decimal.NewFromString(string(bs))
	if e != nil {
		return decimal.Zero, e
	}

	return price, nil
}
func (s *service) GetBorrowRate(ctx context.Context, symbol string, block int64) (decimal.Decimal, error) {
	ratePerBlock, e := s.GetBorrowRatePerBlock(ctx, symbol, block)
	if e != nil {
		return decimal.Zero, e
	}

	return ratePerBlock.Mul(compound.BlocksPerYear), nil
}

func (s *service) SaveSupplyRatePerBlock(ctx context.Context, symbol string, rate decimal.Decimal, block int64) error {
	k := s.supplyRateCacheKey(symbol, block)

	// not exists, add new
	if s.Redis.Exists(k).Val() == 0 {
		//new
		s.Redis.Set(k, []byte(rate.String()), time.Hour)
	}
	return nil
}
func (s *service) GetSupplyRatePerBlock(ctx context.Context, symbol string, block int64) (decimal.Decimal, error) {
	k := s.supplyRateCacheKey(symbol, block)

	bs, e := s.Redis.Get(k).Bytes()
	if e != nil {
		return decimal.Zero, e
	}

	price, e := decimal.NewFromString(string(bs))
	if e != nil {
		return decimal.Zero, e
	}

	return price, nil
}
func (s *service) GetSupplyRate(ctx context.Context, symbol string, block int64) (decimal.Decimal, error) {
	ratePerBlock, e := s.GetSupplyRatePerBlock(ctx, symbol, block)
	if e != nil {
		return decimal.Zero, e
	}

	return ratePerBlock.Mul(compound.BlocksPerYear), nil
}

//资金使用率，同一个block里保持一致，该数据会影响到借款和存款利率的计算
func (s *service) CurUtilizationRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	if market.TotalBorrows.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}

	cash, e := s.mainWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return decimal.Zero, e
	}

	//cash里面不包含准备金，所以不需要减去准备金
	rate := compound.UtilizationRate(cash.Balance, market.TotalBorrows, decimal.Zero)
	return rate, nil
}
func (s *service) CurExchangeRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	if market.CTokens.LessThanOrEqual(decimal.Zero) {
		return market.InitExchangeRate, nil
	}

	cash, e := s.mainWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return decimal.Zero, e
	}

	//cash里面不包含准备金，所以不需要减去准备金
	rate := compound.GetExchangeRate(cash.Balance, market.TotalBorrows, decimal.Zero, market.CTokens, market.InitExchangeRate)

	return rate, nil
}

// 借款年利率
func (s *service) CurBorrowRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	borrowRatePerBlock, e := s.CurBorrowRatePerBlock(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return borrowRatePerBlock.Mul(compound.BlocksPerYear), nil
}

// 借款块利率, 同一个block里保持一致
func (s *service) CurBorrowRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.CurUtilizationRate(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetBorrowRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink)

	return rate, nil
}

// 存款年利率
func (s *service) CurSupplyRate(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	supplyRatePerBlock, e := s.CurSupplyRatePerBlock(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	return supplyRatePerBlock.Mul(compound.BlocksPerYear), nil
}

// 存款块利率, 同一个block里保持一致
func (s *service) CurSupplyRatePerBlock(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	utilRate, e := s.CurUtilizationRate(ctx, market)
	if e != nil {
		return decimal.Zero, e
	}

	rate := compound.GetSupplyRatePerBlock(utilRate, market.BaseRate, market.Multiplier, market.JumpMultiplier, market.Kink, market.ReserveFactor)

	return rate, nil
}

// 剩余现金
func (s *service) CurTotalCash(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	cash, e := s.mainWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return decimal.Zero, e
	}
	return cash.Balance, nil
}

// 总借出量
func (s *service) CurTotalBorrow(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.TotalBorrows, nil
}

// 总保留金
func (s *service) CurTotalReserves(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	cash, e := s.reserveWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return decimal.Zero, e
	}

	return cash.Balance, nil
}

// 总借款利息
func (s *service) CurTotalBorrowInterest(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.TotalBorrowInterest, nil
}

// 总存款利息
func (s *service) CurTotalSupplyInterest(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	return market.TotalSupplyInterest, nil
}

// 铸币
func (s *service) Mint(ctx context.Context, market *core.Market) error {
	return nil
}

func (s *service) utilizationRateCacheKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:urate:%s:%d", symbol, block)
}

func (s *service) borrowRateCacheKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:brate:%s:%d", symbol, block)
}

func (s *service) supplyRateCacheKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:srate:%s:%d", symbol, block)
}
