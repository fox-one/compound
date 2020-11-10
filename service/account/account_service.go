package account

import (
	"compound/core"
	"compound/pkg/id"
	"context"
	"errors"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

type accountService struct {
	mainWallet    *core.Wallet
	marketStore   core.IMarketStore
	supplyStore   core.ISupplyStore
	borrowStore   core.IBorrowStore
	accountStore  core.IAccountStore
	priceService  core.IPriceOracleService
	blockService  core.IBlockService
	walletService core.IWalletService
}

// New new account service
func New(
	mainWallet *core.Wallet,
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
	accountStore core.IAccountStore,
	priceSrv core.IPriceOracleService,
	blockSrv core.IBlockService,
	walletService core.IWalletService,
) core.IAccountService {
	return &accountService{
		mainWallet:    mainWallet,
		marketStore:   marketStore,
		supplyStore:   supplyStore,
		borrowStore:   borrowStore,
		priceService:  priceSrv,
		blockService:  blockSrv,
		walletService: walletService,
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

		value := supply.Principal.Mul(supply.CollateTokens).Div(supply.CTokens).Mul(market.CollateralFactor).Mul(price)
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

func (s *accountService) SeizeTokenAllowed(ctx context.Context, supply *core.Supply, borrow *core.Borrow, seizeTokens decimal.Decimal) bool {
	if supply.UserID != borrow.UserID {
		return false
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return false
	}

	supplyMarket, e := s.marketStore.Find(ctx, "", supply.Symbol)
	if e != nil {
		return false
	}

	maxSeize := supply.CollateTokens.Div(supply.CTokens).Mul(supply.Principal).Mul(supplyMarket.CloseFactor)
	if seizeTokens.GreaterThan(maxSeize) {
		return false
	}

	supplyPrice, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
	if e != nil {
		return false
	}
	borrowPrice, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
	if e != nil {
		return false
	}
	seizePrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
	seizeValue := seizeTokens.Mul(seizePrice)
	borrowValue := borrow.Principal.Mul(borrowPrice)
	if seizeValue.GreaterThan(borrowValue) {
		return false
	}

	return true
}

func (s *accountService) MaxSeize(ctx context.Context, supply *core.Supply, borrow *core.Borrow) (decimal.Decimal, error) {
	if supply.UserID != borrow.UserID {
		return decimal.Zero, errors.New("different user bettween supply and borrow")
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return decimal.Zero, e
	}

	supplyMarket, e := s.marketStore.Find(ctx, "", supply.Symbol)
	if e != nil {
		return decimal.Zero, e
	}

	maxSeize := supply.CollateTokens.Div(supply.CTokens).Mul(supply.Principal).Mul(supplyMarket.CloseFactor)

	supplyPrice, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
	if e != nil {
		return decimal.Zero, e
	}
	borrowPrice, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
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

// SeizeToken  seizeTokens: 可以夺取的币的数量
func (s *accountService) SeizeToken(ctx context.Context, supply *core.Supply, borrow *core.Borrow, seizeTokens decimal.Decimal) (string, error) {
	//同一个block内的同一个人(账号)只允许有一个seize事件，支付borrow的币，夺取supply的币
	if !s.SeizeTokenAllowed(ctx, supply, borrow, seizeTokens) {
		return "", errors.New("seize token not allowed")
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return "", e
	}

	borrowMarket, e := s.marketStore.Find(ctx, "", borrow.Symbol)
	if e != nil {
		return "", e
	}

	supplyPrice, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
	if e != nil {
		return "", e
	}

	borrowPrice, e := s.priceService.GetUnderlyingPrice(ctx, borrow.Symbol, curBlock)
	if e != nil {
		return "", e
	}

	seizeValue := seizeTokens.Mul(supplyPrice)
	repayTokens := seizeValue.Div(borrowPrice)

	trace := id.UUIDFromString(fmt.Sprintf("seizetoken-%s-%d", supply.UserID, curBlock))
	input := mixin.TransferInput{
		AssetID:    borrowMarket.AssetID,
		OpponentID: s.mainWallet.Client.ClientID,
		Amount:     repayTokens.Truncate(8),
		TraceID:    trace,
	}

	if s.walletService.VerifyPayment(ctx, &input) {
		return "", errors.New("seize not allowed")
	}

	action := core.NewAction()
	action[core.ActionKeyService] = core.ActionServiceSeizeToken
	action[core.ActionKeyUser] = supply.UserID
	action[core.ActionKeySymbol] = supply.Symbol

	memoStr, e := action.Format()
	if e != nil {
		return "", e
	}

	input.Memo = memoStr

	return s.walletService.PaySchemaURL(input.Amount, input.AssetID, input.OpponentID, input.TraceID, input.Memo)
}

// SeizeAllowedSupplies get current seize allowed supplies
func (s *accountService) SeizeAllowedSupplies(ctx context.Context) ([]*core.Supply, error) {
	markets, e := s.markets(ctx)
	if e != nil {
		return nil, e
	}

	supplies, e := s.supplyStore.All(ctx)
	if e != nil {
		return nil, e
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return nil, e
	}

	seizeAllowedSupplies := make([]*core.Supply, 0)
	surplusLiquidities := make(map[string]decimal.Decimal)
	for _, supply := range supplies {
		liquidity, found := surplusLiquidities[supply.UserID]
		if !found {
			liquidity, e = s.accountStore.FindLiquidity(ctx, supply.UserID, curBlock)
			if e != nil {
				continue
			}
			// 流动性为0或负的用户加入清算名单
			if liquidity.LessThanOrEqual(decimal.Zero) {
				surplusLiquidities[supply.UserID] = liquidity
			}
		}
		//
		surplus, found := surplusLiquidities[supply.UserID]
		if found {
			market, found := markets[supply.Symbol]
			if !found {
				continue
			}

			price, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
			if e != nil {
				continue
			}

			if surplus.LessThanOrEqual(decimal.Zero) {
				// add to seize allow supply list
				seizeAllowedSupplies = append(seizeAllowedSupplies, supply)

				closeFactor := market.CloseFactor
				liquidationIncentive := market.LiquidationIncentive
				seizePrice := price.Sub(price.Mul(liquidationIncentive))

				maxSeizedValue := supply.CollateTokens.Div(supply.CTokens).Mul(supply.Principal).Mul(seizePrice).Mul(closeFactor)
				delta := surplus.Add(maxSeizedValue)
				surplusLiquidities[supply.UserID] = delta
			}
		}
	}

	return seizeAllowedSupplies, nil
}
