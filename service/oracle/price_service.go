package oracle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"compound/core"
	"compound/pkg/resthttp"

	"github.com/go-redis/redis"
	"github.com/shopspring/decimal"
)

// PriceService price service
type PriceService struct {
	Config       *core.Config
	Redis        *redis.Client
	BlockService core.IBlockService
}

// New new oracle price service
func New(config *core.Config, redis *redis.Client, blockSrv core.IBlockService) core.IPriceOracleService {
	return &PriceService{
		Config:       config,
		Redis:        redis,
		BlockService: blockSrv,
	}
}

// GetUnderlyingPrice get underlying price of asset
func (s *PriceService) GetUnderlyingPrice(ctx context.Context, symbol string, block int64) (decimal.Decimal, error) {
	k := s.redisKey(symbol, block)

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

// Save save price
func (s *PriceService) Save(ctx context.Context, symbol string, price decimal.Decimal, block int64) error {
	k := s.redisKey(symbol, block)

	// not exists, add new
	if s.Redis.Exists(k).Val() == 0 {
		//new expired after 24h
		s.Redis.Set(k, []byte(price.String()), time.Hour*24)
	}

	return nil
}

// PullPriceTicker pull price ticker
func (s *PriceService) PullPriceTicker(ctx context.Context, symbol string, t time.Time) (*core.PriceTicker, error) {
	url := fmt.Sprintf("%s/api/tickers/%s?ts=%d", s.Config.PriceOracle.EndPoint, symbol, t.UTC().Unix())
	resp, err := resthttp.Request(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	var price core.PriceTicker
	err = resthttp.ParseResponse(resp, &price)
	if err != nil {
		return nil, err
	}

	return &price, nil
}

// PullAllPriceTickers pull all price tickers
func (s *PriceService) PullAllPriceTickers(ctx context.Context, t time.Time) ([]*core.PriceTicker, error) {
	url := fmt.Sprintf("%s/api/tickers?ts=%d", s.Config.PriceOracle.EndPoint, t.UTC().Unix())
	resp, err := resthttp.Request(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	var prices []*core.PriceTicker
	err = resthttp.ParseResponse(resp, &prices)
	if err != nil {
		return nil, err
	}

	return prices, nil
}

func (s *PriceService) redisKey(symbol string, block int64) string {
	return fmt.Sprintf("foxone:compound:%s:%d", strings.ToUpper(symbol), block)
}
