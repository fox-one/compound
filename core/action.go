package core

import "encoding/json"

type TransferAction struct {
}

const (
	// ActionKeyService key service type :string
	ActionKeyService = "srv"
	// ActionKeyBlock block index :int64
	ActionKeyBlock = "b"
	// ActionKeySymbol symbol key :string
	ActionKeySymbol = "sb"
	// ActionKeyPrice price :decimal
	ActionKeyPrice = "pr"
	// ActionKeyUtilizationRate utilization rate :decimal
	ActionKeyUtilizationRate = "ur"
	// ActionKeyBorrowRate borrow rate :decimal
	ActionKeyBorrowRate = "br"
	// ActionKeySupplyRate supply rate : decimal
	ActionKeySupplyRate = "sr"
	// ActionKeyAmount amount
	ActionKeyAmount = "amnt"
	// ActionKeyCToken ctokens
	ActionKeyCToken = "ctk"
	// ActionKeyStatus status
	ActionKeyStatus = "st"
	// ActionKeyUser user
	ActionKeyUser = "usr"
)

const (
	// ActionServiceBlock block
	ActionServiceBlock = "blk"
	// ActionServicePrice prc
	ActionServicePrice = "prc"
	// ActionServiceMarket market
	ActionServiceMarket = "mkt"
	// ActionServiceSupply supply
	ActionServiceSupply = "spl"
	// ActionServiceBorrow brw
	ActionServiceBorrow = "brw"
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
