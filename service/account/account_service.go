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
	marketService core.IMarketService
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

func (s *accountService) CalculateAccountLiquidity(ctx context.Context, userID string) (decimal.Decimal, error) {
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
		market, e := s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
		if e != nil {
			continue
		}

		price, e := s.priceService.GetUnderlyingPrice(ctx, market.Symbol, curBlock)
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

func (s *accountService) SeizeTokenAllowed(ctx context.Context, supply *core.Supply, borrow *core.Borrow, repayAmount decimal.Decimal) bool {
	if supply.UserID != borrow.UserID {
		return false
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return false
	}

	supplyMarket, e := s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
	if e != nil {
		return false
	}

	exchangeRate, e := s.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		return false
	}

	maxSeize := supply.Collaterals.Mul(exchangeRate).Mul(supplyMarket.CloseFactor)

	supplyPrice, e := s.priceService.GetUnderlyingPrice(ctx, supplyMarket.Symbol, curBlock)
	if e != nil {
		return false
	}
	borrowMarket, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
	if e != nil {
		return false
	}
	borrowPrice, e := s.priceService.GetUnderlyingPrice(ctx, borrowMarket.Symbol, curBlock)
	if e != nil {
		return false
	}

	repayValue := repayAmount.Mul(borrowPrice)

	seizePrice := supplyPrice.Sub(supplyPrice.Mul(supplyMarket.LiquidationIncentive))
	maxSeizeValue := maxSeize.Mul(seizePrice)

	if repayValue.GreaterThan(maxSeizeValue) {
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

	supplyMarket, e := s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
	if e != nil {
		return decimal.Zero, e
	}

	exchangeRate, e := s.marketService.CurExchangeRate(ctx, supplyMarket)
	if e != nil {
		return decimal.Zero, e
	}

	maxSeize := supply.Collaterals.Mul(exchangeRate).Mul(supplyMarket.CloseFactor)

	supplyPrice, e := s.priceService.GetUnderlyingPrice(ctx, supplyMarket.Symbol, curBlock)
	if e != nil {
		return decimal.Zero, e
	}
	borrowMarket, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
	if e != nil {
		return decimal.Zero, e
	}
	borrowPrice, e := s.priceService.GetUnderlyingPrice(ctx, borrowMarket.Symbol, curBlock)
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
func (s *accountService) SeizeToken(ctx context.Context, supply *core.Supply, borrow *core.Borrow, repayAmount decimal.Decimal) (string, error) {
	//同一个block内的同一个人(账号)只允许有一个seize事件，支付borrow的币，夺取supply的币
	if !s.SeizeTokenAllowed(ctx, supply, borrow, repayAmount) {
		return "", errors.New("seize token not allowed")
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return "", e
	}

	supplyMarket, e := s.marketStore.FindByCToken(ctx, supply.CTokenAssetID)
	if e != nil {
		return "", e
	}

	borrowMarket, e := s.marketStore.FindBySymbol(ctx, borrow.Symbol)
	if e != nil {
		return "", e
	}

	trace := id.UUIDFromString(fmt.Sprintf("seizetoken-%s-%d", supply.UserID, curBlock))
	input := mixin.TransferInput{
		AssetID:    borrowMarket.AssetID,
		OpponentID: s.mainWallet.Client.ClientID,
		Amount:     repayAmount,
		TraceID:    trace,
	}

	if s.walletService.VerifyPayment(ctx, &input) {
		return "", errors.New("seize not allowed")
	}

	action := core.NewAction()
	action[core.ActionKeyService] = core.ActionServiceSeizeToken
	action[core.ActionKeyUser] = supply.UserID
	action[core.ActionKeySymbol] = supplyMarket.Symbol
	action[core.ActionKeyBorrowTrace] = borrow.Trace

	memoStr, e := action.Format()
	if e != nil {
		return "", e
	}

	input.Memo = memoStr

	return s.walletService.PaySchemaURL(input.Amount, input.AssetID, input.OpponentID, input.TraceID, input.Memo)
}

func (s *accountService) SeizeAllowedAccounts(ctx context.Context) ([]*core.Account, error) {
	accounts := make([]*core.Account, 0)

	users, e := s.borrowStore.Users(ctx)
	if e != nil {
		return nil, e
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return nil, e
	}

	for _, u := range users {
		liquidity, e := s.accountStore.FindLiquidity(ctx, u, curBlock)
		if e != nil {
			continue
		}
		if liquidity.LessThanOrEqual(decimal.Zero) {
			supplies, e := s.supplyStore.FindByUser(ctx, u)
			if e != nil {
				continue
			}

			borrows, e := s.borrowStore.FindByUser(ctx, u)
			if e != nil {
				continue
			}
			account := core.Account{
				UserID:    u,
				Liquidity: liquidity,
				Supplies:  supplies,
				Borrows:   borrows,
			}

			accounts = append(accounts, &account)
		}
	}

	return accounts, nil
}
