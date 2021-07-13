package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

const (
	// TransactionKeyService key service type :string
	TransactionKeyService = "service"
	// TransactionKeyBlock block index :int64
	TransactionKeyBlock = "block"
	// TransactionKeySymbol symbol key :string
	TransactionKeySymbol = "symbol"
	// TransactionKeyPrice price :decimal
	TransactionKeyPrice = "price"
	// TransactionKeyBorrowRate borrow rate :decimal
	TransactionKeyBorrowRate = "borrow_rate"
	// TransactionKeySupplyRate supply rate : decimal
	TransactionKeySupplyRate = "supply_rate"
	// TransactionKeyAmount amount
	TransactionKeyAmount = "amount"
	// TransactionKeyCToken ctokens
	TransactionKeyCToken = "ctoken"
	// TransactionKeyInterest interest
	TransactionKeyInterest = "interest"
	// TransactionKeyStatus status
	TransactionKeyStatus = "status"
	// TransactionKeyUser user
	TransactionKeyUser = "user"
	// TransactionKeyErrorCode error code
	TransactionKeyErrorCode = "error_code"
	// TransactionKeyReferTrace refer trace
	TransactionKeyReferTrace = "refer_trace"
	// TransactionKeyAssetID asset id
	TransactionKeyAssetID = "asset_id"
	// TransactionKeyTotalCash total cash
	TransactionKeyTotalCash = "total_cash"
	// TransactionKeyTotalBorrows total borrows
	TransactionKeyTotalBorrows = "total_borrows"
	// TransactionKeyReserves reserves
	TransactionKeyReserves = "reserves"
	// TransactionKeyCTokens ctokens
	TransactionKeyCTokens = "ctokens"
	// TransactionKeyCTokenAssetID ctoken asset id
	TransactionKeyCTokenAssetID = "ctoken_asset_id"
	// TransactionKeyRefund refund
	TransactionKeyRefund = "refund"
	// TransactionKeyOrigin origin
	TransactionKeyOrigin = "origin"
	// TransactionKeySupply supply
	TransactionKeySupply = "supply"
	// TransactionKeyBorrow borrow
	TransactionKeyBorrow = "borrow"
	// TransactionKeyMarket market
	TransactionKeyMarket = "market"
)

type ExtraDataFormatter interface {
	Format() []byte
}

// TransactionExtraData extra data
type TransactionExtraData map[string]interface{}

// NewTransactionExtra new transaction extra instance
func NewTransactionExtra() TransactionExtraData {
	d := make(TransactionExtraData)
	return d
}

// Put put data
func (t TransactionExtraData) Put(key string, value interface{}) {
	t[key] = value
}

func (t TransactionExtraData) GetDecimal(key string) (decimal.Decimal, error) {
	value, ok := t[key].(string)
	if !ok {
		return decimal.Zero, errors.New("type cast error")
	}

	return decimal.NewFromString(value)
}

func (t TransactionExtraData) GetString(key string) (string, error) {
	value, ok := t[key].(string)
	if !ok {
		return "", errors.New("type cast error")
	}

	return value, nil
}

// Format format as []byte by default
func (t TransactionExtraData) Format() []byte {
	bs, e := json.Marshal(t)
	if e != nil {
		return []byte("{}")
	}

	return bs
}

type ExtraSupply struct {
	UserID        string          `json:"user_id,omitempty"`
	CTokenAssetID string          `json:"ctoken_asset_id,omitempty"`
	Collaterals   decimal.Decimal `json:"collaterals,omitempty"`
}

type ExtraBorrow struct {
	UserID        string          `json:"user_id,omitempty"`
	AssetID       string          `json:"asset_id,omitempty"`
	Principal     decimal.Decimal `json:"principal,omitempty"`
	InterestIndex decimal.Decimal `json:"interest_index,omitempty"`
}

type TransactionStatus int

const (
	TransactionStatusInit TransactionStatus = iota
	TransactionStatusComplete
	TransactionStatusAbort
)

// Transaction transaction info
type Transaction struct {
	ID              int64           `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id,omitempty"`
	Action          ActionType      `json:"action,omitempty"`
	TraceID         string          `sql:"size:36;unique_index:idx_transactions_trace_id" json:"trace_id,omitempty"`
	UserID          string          `sql:"size:36;index:idx_transactions_user_id" json:"user_id,omitempty"`
	FollowID        string          `sql:"size:36;index:idx_transactions_follow_id" json:"follow_id,omitempty"`
	SnapshotTraceID string          `sql:"size:36" json:"snapshot_trace_id,omitempty"`
	AssetID         string          `sql:"size:36;index:idx_transactions_asset_id" json:"asset_id,omitempty"`
	Amount          decimal.Decimal `sql:"type:decimal(32,8)" json:"amount,omitempty"`
	Data            types.JSONText  `sql:"type:TEXT" json:"data,omitempty"`
	CreatedAt       time.Time       `sql:"default:CURRENT_TIMESTAMP;index:idx_transactions_created_at" json:"created_at,omitempty"`
	UpdatedAt       time.Time       `sql:"default:CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
}

func (t *Transaction) SetExtraData(extra ExtraDataFormatter) {
	data := []byte("{}")
	if extra != nil {
		data = extra.Format()
	}

	t.Data = data
}

func (t *Transaction) UnmarshalExtraData(obj interface{}) error {
	if obj == nil {
		return errors.New("nil object")
	}

	if err := json.Unmarshal(t.Data, obj); err != nil {
		return err
	}

	return nil
}

// TransactionStore transaction store interface
type TransactionStore interface {
	Create(ctx context.Context, transactions *Transaction) error
	FindByTraceID(ctx context.Context, traceID string) (*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	List(ctx context.Context, offset time.Time, limit int, status TransactionStatus) ([]*Transaction, error)
}

// BuildTransactionFromOutput transaction from output
func BuildTransactionFromOutput(ctx context.Context, userID, followID string, actionType ActionType, output *Output, extra ExtraDataFormatter) *Transaction {
	return &Transaction{
		UserID:   userID,
		Action:   actionType,
		TraceID:  output.TraceID,
		FollowID: followID,
		Amount:   output.Amount,
		AssetID:  output.AssetID,
		Data:     extra.Format(),
	}
}

// BuildTransactionFromTransfer transaction from transfer
func BuildTransactionFromTransfer(ctx context.Context, transfer *Transfer, snapshotTraceID string) (*Transaction, error) {
	var transferAction TransferAction
	m := decodeTransferMemo(transfer.Memo)
	err := json.Unmarshal(m, &transferAction)
	if err != nil {
		return nil, err
	}

	userID := ""
	if len(transfer.Opponents) > 0 {
		userID = transfer.Opponents[0]
	}

	transactionExtra := NewTransactionExtra()
	transactionExtra.Put(TransactionKeyOrigin, transferAction.Origin)
	if transferAction.Code > 0 {
		transactionExtra.Put(TransactionKeyErrorCode, transferAction.Code)
	}

	action := transferAction.Source
	if action == ActionTypeDefault {
		action = ActionTypeProposalTransfer
	}

	return &Transaction{
		UserID:          userID,
		Action:          action,
		TraceID:         transfer.TraceID,
		FollowID:        transferAction.FollowID,
		Amount:          transfer.Amount,
		AssetID:         transfer.AssetID,
		SnapshotTraceID: snapshotTraceID,
		Data:            transactionExtra.Format(),
	}, nil
}

func decodeTransferMemo(memo string) []byte {
	if b, err := base64.StdEncoding.DecodeString(memo); err == nil {
		return b
	}

	if b, err := base64.URLEncoding.DecodeString(memo); err == nil {
		return b
	}

	return []byte(memo)
}
