package core

import "encoding/json"

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
)

const (
	// ActionServicePrice prc
	ActionServicePrice = "prc"
	// ActionServiceSupply supply
	ActionServiceSupply = "spl"
	// ActionServiceBorrow brw
	ActionServiceBorrow = "brw"
	// ActionServiceBorrowTransfer borrow transfer
	ActionServiceBorrowTransfer = "brw-tran"
	// ActionServiceRedeem redeem supply
	ActionServiceRedeem = "rdm"
	// ActionServiceRedeemTransfer redeem transfer to user
	ActionServiceRedeemTransfer = "rdm-tran"
	// ActionServiceRepay repay borrow
	ActionServiceRepay = "rpy"
	// ActionServiceMint mint
	ActionServiceMint = "mint"
	// ActionServicePledge pledge
	ActionServicePledge = "plg"
	// ActionServiceUnpledge unpledge
	ActionServiceUnpledge = "uplg"
	// ActionServiceUnpledgeTransfer unpledge transfer
	ActionServiceUnpledgeTransfer = "uplg-tran"
	// ActionServiceBorrowInterest borrow interest
	ActionServiceBorrowInterest = "b-inter"
	// ActionServiceReserve reserve
	ActionServiceReserve = "reserve"
	// ActionServiceSeizeToken seize token
	ActionServiceSeizeToken = "seize"
	// ActionServiceSeizeTokenTransfer transfer seized token to user
	ActionServiceSeizeTokenTransfer = "seize-tran"
	// ActionServiceRefund refund
	ActionServiceRefund = "rfd"
	// ActionServiceRequestMarket query market
	ActionServiceRequestMarket = "r-mkt"
	// ActionServiceMarketResponse market response
	ActionServiceMarketResponse = "mkt-r"
	// ActionServiceRequestSupply request supply
	ActionServiceRequestSupply = "r-spl"
	// ActionServiceSuppyResponse supply response
	ActionServiceSuppyResponse = "spl-r"
	// ActionServiceRequestBorrow request borrow
	ActionServiceRequestBorrow = "r-brw"
	// ActionServiceBorrowResponse borrow response
	ActionServiceBorrowResponse = "brw-r"
	// ActionServiceRequestLiquidity request liquidity
	ActionServiceRequestLiquidity = "r-lqd"
	// ActionServiceLiquidityResponse liquidity response
	ActionServiceLiquidityResponse = "lqd-r"
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
