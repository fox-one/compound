package account

import (
	"compound/core"
	"context"

	"github.com/shopspring/decimal"
)

type accountService struct {
	marketStore  core.IMarketStore
	supplyStore  core.ISupplyStore
	borrowStore  core.IBorrowStore
	priceService core.IPriceOracleService
	blockService core.IBlockService
}

// New new account service
func New(
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	priceSrv core.IPriceOracleService,
	blockSrv core.IBlockService,
) core.IAccountService {
	return &accountService{
		marketStore:  marketStore,
		supplyStore:  supplyStore,
		borrowStore:  borrowStore,
		priceService: priceSrv,
		blockService: blockSrv,
	}
}

func (s *accountService) CalculateAccountLiquidity(ctx context.Context, userID string) (decimal.Decimal, error) {
	markets, e := s.markets(ctx)
	if e != nil {
		return decimal.Zero, e
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return decimal.Zero, e
	}

	supplies, e := s.supplyStore.FindByUser(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}
	supplyValue := decimal.Zero
	for _, supply := range supplies {
		market, found := markets[supply.Symbol]
		if !found {
			continue
		}

		price, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
		if e != nil {
			continue
		}

		value := supply.Principal.Mul(supply.CollateTokens).Div(supply.Ctokens).Mul(market.CollateralFactor).Mul(price)
		supplyValue = supplyValue.Add(value)
	}

	borrows, e := s.borrowStore.FindByUser(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}

	borrowValue := decimal.Zero

	for _, borrow := range borrows {
		price, e := s.priceService.GetUnderlyingPrice(ctx, borrow.Symbol, curBlock)
		if e != nil {
			continue
		}

		value := borrow.Principal.Mul(price)
		borrowValue = borrowValue.Add(value)
	}

	liquidity := supplyValue.Sub(borrowValue)

	return liquidity, nil
}

func (s *accountService) markets(ctx context.Context) (map[string]*core.Market, error) {
	markets, e := s.marketStore.All(ctx)
	if e != nil {
		return nil, e
	}

	maps := make(map[string]*core.Market)

	for _, m := range markets {
		maps[m.Symbol] = m
	}

	return maps, nil
}

func (s *accountService) HasBorrows(ctx context.Context, userID string) (bool, error) {
	borrows, e := s.borrowStore.FindByUser(ctx, userID)
	if e != nil {
		return false, e
	}

	for _, b := range borrows {
		if b.Principal.GreaterThan(decimal.Zero) {
			return true, nil
		}
	}

	return false, nil
}
