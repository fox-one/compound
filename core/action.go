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
	// ActionTypeSeizeToken seize token
	ActionTypeSeizeToken
	// ActionTypeRedeemTransfer redeem transfer
	ActionTypeRedeemTransfer
	// ActionTypeUnpledgeTransfer unpledge transfer
	ActionTypeUnpledgeTransfer
	// ActionTypeBorrowTransfer borrow transfer
	ActionTypeBorrowTransfer
	// ActionTypeSeizeTokenTransfer seize token transfer
	ActionTypeSeizeTokenTransfer
	// ActionTypeRefundTransfer refund
	ActionTypeRefundTransfer
	// ActionTypeRepayRefundTransfer repay refund
	ActionTypeRepayRefundTransfer
	// ActionTypeSeizeRefundTransfer seize refund
	ActionTypeSeizeRefundTransfer
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
)
