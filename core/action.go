package core

import (
	"compound/pkg/mtg/types"

	"github.com/fox-one/msgpack"
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
	// ActionTypeProposalUpsertMarket add market proposal action
	ActionTypeProposalUpsertMarket
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
	_
	_
	_
	_
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

func (a ActionType) IsProposalAction() bool {
	return a == ActionTypeProposalUpsertMarket ||
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
		a == ActionTypeProposalAddOracleSigner ||
		a == ActionTypeProposalRemoveOracleSigner ||
		a == ActionTypeProposalSetProperty
}

func (i ActionType) MarshalBinary() (data []byte, err error) {
	return types.BitInt(i).MarshalBinary()
}

func (i *ActionType) UnmarshalBinary(data []byte) error {
	var b types.BitInt
	if err := b.UnmarshalBinary(data); err != nil {
		return err
	}

	*i = ActionType(b)
	return nil
}

type TransactionAction struct {
	FollowID []byte `msgpack:"f,omitempty"`
	Body     []byte `msgpack:"b,omitempty"`
}

func (action TransactionAction) Encode() ([]byte, error) {
	return msgpack.Marshal(action)
}

func DecodeTransactionAction(b []byte) (*TransactionAction, error) {
	var action TransactionAction
	if err := msgpack.Unmarshal(b, &action); err != nil {
		return nil, err
	}

	return &action, nil
}
