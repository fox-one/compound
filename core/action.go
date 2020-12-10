package core

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

//go:generate stringer -type ActionType -trimprefix ActionType

// ActionType compound action type
type ActionType int

const (
	_ ActionType = iota
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
	// ActionTypeProposalAddMarket add market
	ActionTypeProposalAddMarket
	// ActionTypeProposalUpdateMarket update market
	ActionTypeProposalUpdateMarket
	// ActionTypeProposalInjectCTokenForMint inject token
	ActionTypeProposalInjectCTokenForMint
	// ActionTypeProposalWithdrawReserves withdraw
	ActionTypeProposalWithdrawReserves
	// ActionTypeProposalProvidePrice price
	ActionTypeProposalProvidePrice
	// ActionTypeProposalVote vote
	ActionTypeProposalVote
)
