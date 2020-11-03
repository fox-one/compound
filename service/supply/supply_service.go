package supply

import (
	"compound/core"
	"compound/pkg/id"
	"compound/service/wallet"
	"context"
	"errors"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/store/db"
	"github.com/shopspring/decimal"
)

type supplyService struct {
	config         *core.Config
	mainWallet     *mixin.Client
	blockWallet    *mixin.Client
	db             *db.DB
	supplyStore    core.ISupplyStore
	accountService core.IAccountService
	priceService   core.IPriceOracleService
	blockService   core.IBlockService
}

// New new supply service
func New(cfg *core.Config,
	mainWallet *mixin.Client,
	db *db.DB,
	supplyStore core.ISupplyStore,
	accountService core.IAccountService,
	priceService core.IPriceOracleService,
	blockService core.IBlockService) core.ISupplyService {
	// return &supplyService{
	// 	config:      cfg,
	// 	mainWallet:  mainWallet,
	// 	db:          db,
	// 	supplyStore: supplyStore,
	// }
	return nil
}

// 赎回
func (s *supplyService) Redeem(ctx context.Context, redeemTokens decimal.Decimal, userID string, market *core.Market) (string, error) {
	if !s.RedeemAllowed(ctx, redeemTokens, userID, market) {
		return "", errors.New("redeem not allowed")
	}

	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceRedeem

	str, e := action.Format()
	if e != nil {
		return "", e
	}

	return wallet.PaySchemaURL(redeemTokens, market.CTokenAssetID, s.mainWallet.ClientID, id.GenTraceID(), str)
}

func (s *supplyService) RedeemAllowed(ctx context.Context, redeemTokens decimal.Decimal, userID string, market *core.Market) bool {
	supply, e := s.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return false
	}

	// check ctokens
	if redeemTokens.GreaterThan(supply.Ctokens) {
		return false
	}

	amount := supply.Principal.Mul(redeemTokens).Div(supply.Ctokens)

	// check market cash
	marketCash, e := s.mainWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return false
	}

	if amount.GreaterThan(marketCash.Balance) {
		return false
	}

	if !s.accountService.IsNoBorrows(ctx, userID) &&
		supply.CollateStatus == core.CollateStateOn {
		// check liquidity
		liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID)
		if e != nil {
			return false
		}

		if liquidity.LessThanOrEqual(decimal.Zero) {
			return false
		}

		comp, e := s.accountService.CompValueAndLiquidity(ctx, amount, market.Symbol, liquidity)
		if e != nil {
			return false
		}

		if comp.GreaterThan(decimal.Zero) {
			return false
		}
	}

	return true
}

// return max redeem ctokens
func (s *supplyService) MaxRedeem(ctx context.Context, userID string, market *core.Market) (decimal.Decimal, error) {
	supply, e := s.supplyStore.Find(ctx, userID, market.Symbol)
	if e != nil {
		return decimal.Zero, e
	}

	amount := supply.Principal

	if !s.accountService.IsNoBorrows(ctx, userID) &&
		supply.CollateStatus == core.CollateStateOn {
		// check liquidity
		liquidity, e := s.accountService.CalculateAccountLiquidity(ctx, userID)
		if e != nil {
			return decimal.Zero, e
		}

		if liquidity.LessThanOrEqual(decimal.Zero) {
			return decimal.Zero, e
		}

		curBlock, e := s.blockService.CurrentBlock(ctx)
		if e != nil {
			return decimal.Zero, e
		}

		price, e := s.priceService.GetUnderlyingPrice(ctx, market.Symbol, curBlock)
		if e != nil {
			return decimal.Zero, e
		}

		redeemValue := amount.Mul(price)
		if redeemValue.GreaterThan(liquidity) {
			redeemValue = liquidity
		}

		amount = redeemValue.Div(price)
	}

	// check market cash
	marketCash, e := s.mainWallet.ReadAsset(ctx, market.AssetID)
	if e != nil {
		return decimal.Zero, e
	}

	if amount.GreaterThan(marketCash.Balance) {
		amount = marketCash.Balance
	}

	ctokens := supply.Ctokens.Mul(amount).Div(supply.Principal)

	return ctokens, nil
}

// 存入
func (s *supplyService) Supply(ctx context.Context, amount decimal.Decimal, market *core.Market) (string, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return "", errors.New("invalid amount")
	}

	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceSupply

	str, e := action.Format()
	if e != nil {
		return "", e
	}
	return wallet.PaySchemaURL(amount, market.AssetID, s.mainWallet.ClientID, id.GenTraceID(), str)
}

func (s *supplyService) SetCollateStatus(ctx context.Context, userID string, market *core.Market, status core.CollateStatus) error {
	log := logger.FromContext(ctx)
	//不可抵押
	if market.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		return errors.New("disable collateral")
	}

	curBlock, e := s.blockService.CurrentBlock(ctx)
	if e != nil {
		return e
	}

	trace := id.UUIDFromString(fmt.Sprintf("collate-%s-%s-%d-%d", userID, market.Symbol, curBlock, status))
	input := mixin.TransferInput{
		AssetID:    s.config.App.BlockAssetID,
		OpponentID: s.mainWallet.ClientID,
		Amount:     decimal.NewFromFloat(0.00000001),
		TraceID:    trace,
	}

	payment, err := s.mainWallet.VerifyPayment(ctx, input)
	if err != nil {
		log.Errorln("verifypayment error:", err)
		return err
	}

	if payment.Status == "paid" {
		log.Infoln("transaction exists")
		return errors.New("transaction exists")
	}

	memo := make(core.Action)
	memo[core.ActionKeyService] = core.ActionServiceCollateralStatus
	memo[core.ActionKeyStatus] = status.String()
	memo[core.ActionKeySymbol] = market.Symbol
	memo[core.ActionKeyUser] = userID

	memoStr, e := s.blockService.FormatBlockMemo(ctx, memo)
	if e != nil {
		return e
	}

	input.Memo = memoStr

	_, e = s.blockWallet.Transfer(ctx, &input, s.config.BlockWallet.Pin)
	if e != nil {
		return e
	}

	return nil

}

// estimatedAccountLiquidity(
// calculateAccountLiquidity

//supplyBalance
//borrowBalance
