package oracle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"compound/core"
	"compound/pkg/resthttp"

	"github.com/fox-one/pkg/logger"
	"github.com/shopspring/decimal"
)

// PriceService price service
type PriceService struct {
	Config       *core.Config
	BlockService core.IBlockService
}

// New new oracle price service
func New(config *core.Config, blockSrv core.IBlockService) core.IPriceOracleService {
	return &PriceService{
		Config:       config,
		BlockService: blockSrv,
	}
}

// GetCurrentUnderlyingPrice get current price of market
func (s *PriceService) GetCurrentUnderlyingPrice(ctx context.Context, market *core.Market) (decimal.Decimal, error) {
	if market.Price.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, errors.New("invalid market price")
	}

	return market.Price, nil
}

// PullPriceTicker pull price ticker
func (s *PriceService) PullPriceTicker(ctx context.Context, assetID string, t time.Time) (*core.PriceTicker, error) {
	url := fmt.Sprintf("%s/api/v2/tickers/%s?ts=%d", s.Config.PriceOracle.EndPoint, assetID, t.UTC().Unix())
	logger.FromContext(ctx).Infoln("pull price:", url)
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
