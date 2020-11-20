package account

import (
	"compound/core"
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type accountService struct {
	mainWallet    *core.Wallet
	marketStore   core.IMarketStore
	supplyStore   core.ISupplyStore
	borrowStore   core.IBorrowStore
	priceService  core.IPriceOracleService
	blockService  core.IBlockService
	walletService core.IWalletService
	marketService core.IMarketService
}

// New new account service
func New(
	mainWallet *core.Wallet,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	priceSrv core.IPriceOracleService,
	blockSrv core.IBlockService,
	walletService core.IWalletService,
	marketServie core.IMarketService,
) core.IAccountService {
	return &accountService{
		mainWallet:    mainWallet,
		marketStore:   marketStore,
		supplyStore:   supplyStore,
		borrowStore:   borrowStore,
		priceService:  priceSrv,
		blockService:  blockSrv,
		walletService: walletService,
		marketService: marketServie,
	}
}

func (s *accountService) CalculateAccountLiquidity(ctx context.Context, userID string, blockNum int64) (decimal.Decimal, error) {
	supplies, e := s.supplyStore.FindByUser(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}
	supplyValue := decimal.Zero
	for _, supply := range supplies {
		market, e := s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
		if e != nil {
			continue
		}

		price, e := s.priceService.GetCurrentUnderlyingPrice(ctx, market)
		if e != nil {
			continue
		}

		exchangeRate, e := s.marketService.CurExchangeRate(ctx, market)
		if e != nil {
			continue
		}
		value := supply.Collaterals.Mul(exchangeRate).Mul(market.CollateralFactor).Mul(price)
		supplyValue = supplyValue.Add(value)
	}

	borrows, e := s.borrowStore.FindByUser(ctx, userID)
	if e != nil {
		return decimal.Zero, e
	}

	borrowValue := decimal.Zero

	for _, borrow := range borrows {
		market, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
		if e != nil {
			continue
		}
		price, e := s.priceService.GetCurrentUnderlyingPrice(ctx, market)
		if e != nil {
			continue
		}

		borrowBalance, e := borrow.Balance(ctx, market)
		if e != nil {
			continue
		}
		value := borrowBalance.Mul(price)
		borrowValue = borrowValue.Add(value)
	}

	liquidity := supplyValue.Sub(borrowValue)

	return liquidity, nil
}

// key with asset_id
func (s *accountService) markets(ctx context.Context) (map[string]*core.Market, error) {
	markets, e := s.marketStore.All(ctx)
	if e != nil {
		return nil, e
	}

	maps := make(map[string]*core.Market)

	for _, m := range markets {
		maps[m.AssetID] = m
	}

	return maps, nil
}

func (s *accountService) SeizeTokenAllowed(ctx context.Context, supply *core.Supply, borrow *core.Borrow, repayAmount decimal.Decimal, time time.Time) bool {
	if supply.UserID != borrow.UserID {
		return false
	}

	blockNum, e := s.blockService.GetBlock(ctx, time)
	if e != nil {
		return false
	}

	// check liquidity
	liquidity, e := s.CalculateAccountLiquidity(ctx, borrow.UserID, blockNum)
	if e != nil {
		return false
	}

	if liquidity.GreaterThan(decimal.Zero) {
		return false
	}

	return true
}

func (s *accountService) MaxSeize(ctx context.Context, supply *core.Supply, borrow *core.Borrow) (decimal.Decimal, error) {
	if supply.UserID != borrow.UserID {
		return decimal.Zero, errors.New("different user bettween supply and borrow")
	}

	supplyMarket, e := s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
	if e != nil {
		return decimal.Zero, e
	}

	exchangeRate, e := s.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		return decimal.Zero, e
	}

	maxSeize := supply.Collaterals.Mul(exchangeRate).Mul(supplyMarket.CloseFactor)

	supplyPrice, e := s.priceService.GetCurrentUnderlyingPrice(ctx, supplyMarket)
	if e != nil {
		return decimal.Zero, e
	}
	borrowMarket, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
	if e != nil {
		return decimal.Zero, e
	}
	borrowPrice, e := s.priceService.GetCurrentUnderlyingPrice(ctx, borrowMarket)
	if e != nil {
		return decimal.Zero, e
	}
	seizePrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
	seizeValue := maxSeize.Mul(seizePrice)
	borrowValue := borrow.Principal.Mul(borrowPrice)
	if seizeValue.GreaterThan(borrowValue) {
		seizeValue = borrowValue
		maxSeize = seizeValue.Div(seizePrice)
	}

	return maxSeize, nil
}

func (s *accountService) SeizeToken(ctx context.Context, supply *core.Supply, borrow *core.Borrow, repayAmount decimal.Decimal) (string, error) {
	return "", nil
}
