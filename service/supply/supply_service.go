package supply

import (
	"compound/core"
	"compound/pkg/id"
	"compound/service/wallet"
	"context"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
)

type supplyService struct {
	config     *core.Config
	mainWallet *mixin.Client
}

// New new supply service
func New(cfg *core.Config, mainWallet *mixin.Client) core.ISupplyService {
	return &supplyService{
		config:     cfg,
		mainWallet: mainWallet,
	}
}

func (s *supplyService) Redeem(ctx context.Context, amount decimal.Decimal, market *core.Market) (string, error) {
	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceRedeem

	str, e := action.Format()
	if e != nil {
		return "", e
	}

	return wallet.PaySchemaURL(amount, market.CTokenAssetID, s.mainWallet.ClientID, id.GenTraceID(), str)
}

func (s *supplyService) Loan(ctx context.Context, amount decimal.Decimal, market *core.Market) (string, error) {
	action := make(core.Action)
	action[core.ActionKeyService] = core.ActionServiceSupply

	str, e := action.Format()
	if e != nil {
		return "", e
	}
	return wallet.PaySchemaURL(amount, market.AssetID, s.mainWallet.ClientID, id.GenTraceID(), str)
}
