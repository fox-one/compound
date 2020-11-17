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
)

const (
	// ActionServicePrice prc
	ActionServicePrice = "prc"
	// ActionServiceMarket market
	ActionServiceMarket = "mkt"
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
	// ActionServiceLiquidity liquidity
	ActionServiceLiquidity = "lqdy"
	// ActionServiceSeizeToken seize token
	ActionServiceSeizeToken = "seize"
	// ActionServiceSeizeTokenTransfer transfer seized token to user
	ActionServiceSeizeTokenTransfer = "seize-tran"
	// ActionServiceRefund refund
	ActionServiceRefund = "rfd"
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
