package borrow

import (
	"compound/core"
	"context"
	"errors"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

type borrowService struct {
	config       *core.Config
	mainWallet   *mixin.Client
	marketStore  core.IMarketStore
	supplyStore  core.ISupplyStore
	blockService core.IBlockService
	priceService core.IPriceOracleService
}

// New new borrow service
func New(cfg *core.Config, mainWallet *mixin.Client, marketStore core.IMarketStore, supplyStore core.ISupplyStore, blockService core.IBlockService, priceService core.IPriceOracleService) core.IBorrowService {
	// return &borrowService{
	// 	config:     cfg,
	// 	mainWallet: mainWallet,
	// }
	return nil
}

func (s *borrowService) Repay(ctx context.Context, amount decimal.Decimal, userID string, market *core.Market) error {
	return nil
}

func (s *borrowService) Borrow(ctx context.Context, borrowAmount decimal.Decimal, borrowAssetID string, collateralAssetID string, userID string) error {
	if borrowAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("invalid borrow amount")
	}

	borrowMarket, e := s.marketStore.Find(ctx, borrowAssetID, "")
	if e != nil {
		return e
	}

	collateralMarket, e := s.marketStore.Find(ctx, collateralAssetID, "")
	if e != nil {
		return e
	}

	// check collateral factor
	if collateralMarket.CollateralFactor.LessThanOrEqual(decimal.Zero) {
		return core.ErrCollateralDisabled
	}

	// check borrow cap
	borrowMarketCash, e := s.mainWallet.ReadAsset(ctx, borrowMarket.AssetID)
	if e != nil {
		return e
	}
	sub := borrowMarketCash.Balance.Sub(borrowAmount)
	if sub.LessThanOrEqual(borrowMarket.BorrowCap) {
		return core.ErrBorrowsOverCap
	}

	//check liquidity

	return nil
}

// 预估账户的流动性
func (s *borrowService) estimatedAccountLiquidity(ctx context.Context, collateralAmount decimal.Decimal, collateralMarket *core.Market, userID string) (decimal.Decimal, error) {
	// suppies, e := s.supplyStore.Find(ctx, userID)
	// if e != nil {
	// 	return decimal.Zero, e
	// }

	// curBlock, e := s.blockService.CurrentBlock(ctx)
	// if e != nil {
	// 	return decimal.Zero, e
	// }
	// //总可抵押价值
	// totalBorrowablePrice := decimal.Zero
	// for _, supply := range suppies {
	// 	p, e := s.priceService.GetUnderlyingPrice(ctx, supply.Symbol, curBlock)
	// 	if e != nil {
	// 		return decimal.Zero, e
	// 	}

	// 	delta := supply.CTokens.Sub(supply.Collaterals)
	// 	totalBorrowablePrice = totalBorrowablePrice.Add(delta.Mul(p))
	// }

	// if totalBorrowablePrice.LessThanOrEqual(decimal.Zero) {
	// 	return decimal.Zero, nil
	// }

	// bp, e := s.priceService.GetUnderlyingPrice(ctx, collateralMarket.Symbol, curBlock)
	// if e != nil {
	// 	return decimal.Zero, e
	// }

	// collateralPrice := borrowAmount.Mul(bp)

	return decimal.Zero, nil
}
