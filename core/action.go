package core

import (
	"strings"
)

//go:generate stringer -type ActionType -trimprefix ActionType

// ActionType compound action type
type ActionType int

const (
	// ActionTypeDefault default
	ActionTypeDefault ActionType = iota
	// ActionTypeSupply supply action
	ActionTypeSupply
	// ActionTypeBorrow borrow action
	ActionTypeBorrow
	// ActionTypeRedeem redeem action
	ActionTypeRedeem
	// ActionTypeRepay repay action
	ActionTypeRepay
	// ActionTypeMint mint ctoken action
	ActionTypeMint
	// ActionTypePledge pledge action
	ActionTypePledge
	// ActionTypeUnpledge unpledge action
	ActionTypeUnpledge
	// ActionTypeLiquidate liquidation action
	ActionTypeLiquidate
	// ActionTypeRedeemTransfer redeem transfer action
	ActionTypeRedeemTransfer
	// ActionTypeUnpledgeTransfer unpledge transfer action
	ActionTypeUnpledgeTransfer
	// ActionTypeBorrowTransfer borrow transfer action
	ActionTypeBorrowTransfer
	// ActionTypeLiquidateTransfer liquidation transfer action
	ActionTypeLiquidateTransfer
	// ActionTypeRefundTransfer refund action
	ActionTypeRefundTransfer
	// ActionTypeRepayRefundTransfer repay refund action
	ActionTypeRepayRefundTransfer
	// ActionTypeLiquidateRefundTransfer seize refund action
	ActionTypeLiquidateRefundTransfer
	// ActionTypeProposalAddMarket add market proposal action
	ActionTypeProposalAddMarket
	// ActionTypeProposalUpdateMarket update market proposal action
	ActionTypeProposalUpdateMarket
	// ActionTypeProposalWithdrawReserves withdraw reserves proposal action
	ActionTypeProposalWithdrawReserves
	// ActionTypeProposalProvidePrice provide price action
	ActionTypeProposalProvidePrice
	// ActionTypeProposalVote vote action
	ActionTypeProposalVote
	// ActionTypeProposalInjectCTokenForMint inject token action
	ActionTypeProposalInjectCTokenForMint
	// ActionTypeProposalUpdateMarketAdvance update market advance parameters action
	ActionTypeProposalUpdateMarketAdvance
	// ActionTypeProposalTransfer proposal transfer action
	ActionTypeProposalTransfer
	// ActionTypeProposalCloseMarket proposal close market action
	ActionTypeProposalCloseMarket
	// ActionTypeProposalOpenMarket proposal open market action
	ActionTypeProposalOpenMarket
	// ActionTypeProposalAddScope proposal add allowlist scope action
	ActionTypeProposalAddScope
	// ActionTypeProposalRemoveScope proposal remove allowlist scope action
	ActionTypeProposalRemoveScope
	// ActionTypeProposalAddAllowList proposal add to allowlist action
	ActionTypeProposalAddAllowList
	// ActionTypeProposalRemoveAllowList proposal remove from allowlist action
	ActionTypeProposalRemoveAllowList
	// ActionTypeUpdateMarket update market
	ActionTypeUpdateMarket
	// ActionTypeQuickPledge supply -> pledge
	ActionTypeQuickPledge
	// ActionTypeQuickBorrow supply -> pledge -> borrow
	ActionTypeQuickBorrow
	// ActionTypeQuickBorrowTransfer quick borrow transfer
	ActionTypeQuickBorrowTransfer
	// ActionTypeQuickRedeem unpledge -> redeem
	ActionTypeQuickRedeem
	// ActionTypeQuickRedeem quick redeem transfer
	ActionTypeQuickRedeemTransfer
	// ActionTypeProposalAddOracleSigner add oracle signer proposal action
	ActionTypeProposalAddOracleSigner
	// ActionTypeProposalRemoveOracleSigner remove oracle signer proposal action
	ActionTypeProposalRemoveOracleSigner
	// ActionTypeProposalSetProperty proposal to set property value
	ActionTypeProposalSetProperty
	ActionTypeProposalMake
	ActionTypeProposalShout
)

func ParseActionType(t string) ActionType {
	for idx := 0; idx < len(_ActionType_index)-1; idx++ {
		l, r := _ActionType_index[idx], _ActionType_index[idx+1]
		if typ := _ActionType_name[l:r]; strings.EqualFold(typ, t) {
			return ActionType(idx)
		}
	}

	return 0
}

func (a ActionType) IsProposalAction() bool {
	return a == ActionTypeProposalAddMarket ||
		a == ActionTypeProposalUpdateMarket ||
		a == ActionTypeProposalWithdrawReserves ||
		a == ActionTypeProposalProvidePrice ||
		a == ActionTypeProposalVote ||
		a == ActionTypeProposalMake ||
		a == ActionTypeProposalShout ||
		a == ActionTypeProposalInjectCTokenForMint ||
		a == ActionTypeProposalUpdateMarketAdvance ||
		a == ActionTypeProposalCloseMarket ||
		a == ActionTypeProposalOpenMarket ||
		a == ActionTypeProposalAddScope ||
		a == ActionTypeProposalRemoveScope ||
		a == ActionTypeProposalAddAllowList ||
		a == ActionTypeProposalRemoveAllowList ||
		a == ActionTypeProposalAddOracleSigner ||
		a == ActionTypeProposalRemoveOracleSigner ||
		a == ActionTypeProposalSetProperty
}
