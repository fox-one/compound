package core

import (
	"encoding/json"
)

const (
	// ActionKeyService key service type :string
	ActionKeyService = "srv"
	// ActionKeyBlock block index :int64
	ActionKeyBlock = "b"
	// ActionKeySymbol symbol key :string
	ActionKeySymbol = "sb"
	// ActionKeyPrice price :decimal
	ActionKeyPrice = "pr"
	// ActionKeyBorrowRate borrow rate :decimal
	ActionKeyBorrowRate = "br"
	// ActionKeySupplyRate supply rate : decimal
	ActionKeySupplyRate = "sr"
	// ActionKeyAmount amount
	ActionKeyAmount = "amnt"
	// ActionKeyCToken ctokens
	ActionKeyCToken = "ctk"
	// ActionKeyInterest interest
	ActionKeyInterest = "inter"
	// ActionKeyStatus status
	ActionKeyStatus = "st"
	// ActionKeyUser user
	ActionKeyUser = "usr"
	// ActionKeyErrorCode error code
	ActionKeyErrorCode = "ec"
	// ActionKeyReferTrace refer trace
	ActionKeyReferTrace = "rftr"
	// ActionKeyAssetID asset id
	ActionKeyAssetID = "aid"
	// ActionKeyTotalCash total cash
	ActionKeyTotalCash = "tc"
	// ActionKeyTotalBorrows total borrows
	ActionKeyTotalBorrows = "tbs"
	// ActionKeyReserves reserves
	ActionKeyReserves = "rvs"
	// ActionKeyCTokens ctokens
	ActionKeyCTokens = "ctks"
	// ActionKeyCTokenAssetID ctoken asset id
	ActionKeyCTokenAssetID = "ctid"
	// ActionKeyInitExchangeRate exchange rate
	ActionKeyInitExchangeRate = "ier"
	// ActionKeyReserveFactor reserve factor
	ActionKeyReserveFactor = "rf"
	// ActionKeyLiquidationIncentive liquidation incentive
	ActionKeyLiquidationIncentive = "lqi"
	// ActionKeyBorrowCap borrow cap
	ActionKeyBorrowCap = "bc"
	// ActionKeyCollateralFactor collateral factor
	ActionKeyCollateralFactor = "cf"
	// ActionKeyCloseFactor close factor
	ActionKeyCloseFactor = "clf"
	// ActionKeyBaseRate base rate
	ActionKeyBaseRate = "brt"
	// ActionKeyMultiPlier mutliplier
	ActionKeyMultiPlier = "mp"
	// ActionKeyJumpMultiPlier jump multiplier
	ActionKeyJumpMultiPlier = "jmp"
	// ActionKeyKink kink
	ActionKeyKink = "kk"
	// ActionKeyUtilizationRate utilization rate
	ActionKeyUtilizationRate = "ur"
	// ActionKeyExchangeRate exchange rate
	ActionKeyExchangeRate = "er"
	// ActionKeyBorowIndex borrow index
	ActionKeyBorowIndex = "bi"
	// ActionKeyLiquidity liquidity
	ActionKeyLiquidity = "lq"
	// ActionKeyInterestIndex interest index
	ActionKeyInterestIndex = "ii"
	// ActionKeyBorrowBalance borrow balance
	ActionKeyBorrowBalance = "bb"
)

// Action action
type Action map[string]string

// Format format to string
func (a *Action) Format() (string, error) {
	bs, e := json.Marshal(a)
	if e != nil {
		return "", e
	}

	return string(bs), nil
}

// NewAction new action
func NewAction() Action {
	return make(Action)
}

//go:generate stringer -type ActionType -trimprefix ActionType

// ActionType compound action type
type ActionType int

const (
	_ ActionType = iota
	// ActionTypeSupply supply
	ActionTypeSupply
	// ActionTypeBorrow borrow
	ActionTypeBorrow
	// ActionTypeBorrowTransfer borrow transfer
	ActionTypeBorrowTransfer
	// ActionTypeRedeem redeem
	ActionTypeRedeem
	// ActionTypeRedeemTransfer redeem transfer
	ActionTypeRedeemTransfer
	// ActionTypeRepay repay
	ActionTypeRepay
	// ActionTypeMint mint
	ActionTypeMint
	// ActionTypePledge pledge
	ActionTypePledge
	// ActionTypeUnpledge unpledge
	ActionTypeUnpledge
	// ActionTypeUnpledgeTransfer unpledge transfer
	ActionTypeUnpledgeTransfer
	// ActionTypeSeizeToken seize token
	ActionTypeSeizeToken
	// ActionTypeSeizeTokenTransfer seize token transfer
	ActionTypeSeizeTokenTransfer
	// ActionTypeRefund refund
	ActionTypeRefund
	// ActionTypeAddMarket add market
	ActionTypeAddMarket
	// ActionTypeUpdateMarket update market
	ActionTypeUpdateMarket
	// ActionTypeInjectMintToken inject token
	ActionTypeInjectMintToken
	// ActionTypeProposalWithdraw withdraw
	ActionTypeProposalWithdraw
	// ActionTypeProposalPrice price
	ActionTypeProposalPrice
)
