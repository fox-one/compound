package core

//go:generate stringer -type ActionType -trimprefix ActionType

// ActionType compound action type
type ActionType int

const (
	// ActionTypeDefault default
	ActionTypeDefault ActionType = iota
	// ActionTypeSupply supply
	ActionTypeSupply
	// ActionTypeBorrow borrow
	ActionTypeBorrow
	// ActionTypeRedeem redeem
	ActionTypeRedeem
	// ActionTypeRepay repay
	ActionTypeRepay
	// ActionTypeMint mint
	ActionTypeMint
	// ActionTypePledge pledge
	ActionTypePledge
	// ActionTypeUnpledge unpledge
	ActionTypeUnpledge
	// ActionTypeLiquidate seize token
	ActionTypeLiquidate
	// ActionTypeRedeemTransfer redeem transfer
	ActionTypeRedeemTransfer
	// ActionTypeUnpledgeTransfer unpledge transfer
	ActionTypeUnpledgeTransfer
	// ActionTypeBorrowTransfer borrow transfer
	ActionTypeBorrowTransfer
	// ActionTypeLiquidateTransfer seize token transfer
	ActionTypeLiquidateTransfer
	// ActionTypeRefundTransfer refund
	ActionTypeRefundTransfer
	// ActionTypeRepayRefundTransfer repay refund
	ActionTypeRepayRefundTransfer
	// ActionTypeLiquidateRefundTransfer seize refund
	ActionTypeLiquidateRefundTransfer
	// ActionTypeProposalAddMarket add market
	ActionTypeProposalAddMarket
	// ActionTypeProposalUpdateMarket update market
	ActionTypeProposalUpdateMarket
	// ActionTypeProposalWithdrawReserves withdraw
	ActionTypeProposalWithdrawReserves
	// ActionTypeProposalProvidePrice price
	ActionTypeProposalProvidePrice
	// ActionTypeProposalVote vote
	ActionTypeProposalVote
	// ActionTypeProposalInjectCTokenForMint inject token
	ActionTypeProposalInjectCTokenForMint
	// ActionTypeProposalUpdateMarketAdvance update market advance
	ActionTypeProposalUpdateMarketAdvance
	// ActionTypeProposalTransfer proposal transfer
	ActionTypeProposalTransfer
	// ActionTypeProposalCloseMarket proposal close market
	ActionTypeProposalCloseMarket
	// ActionTypeProposalOpenMarket proposal open market
	ActionTypeProposalOpenMarket
	// ActionTypeProposalAddScope proposal add allowlist scope
	ActionTypeProposalAddScope
	// ActionTypeProposalRemoveScope proposal remove allowlist scope
	ActionTypeProposalRemoveScope
	// ActionTypeProposalAddAllowList proposal add to allowlist
	ActionTypeProposalAddAllowList
	// ActionTypeProposalRemoveAllowList proposal remove from allowlist
	ActionTypeProposalRemoveAllowList
)
