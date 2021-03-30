package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	uuidutil "github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

// TransferAction  transfer action
type TransferAction struct {
	Code     int        `json:"c,omitempty"`
	Origin   ActionType `json:"o,omitempty"`
	Source   ActionType `json:"s,omitempty"`
	FollowID string     `json:"f,omitempty"`
	Message  string     `json:"m,omitempty"`
}

// Format format TransferAction to string
func (t *TransferAction) Format() (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func NewRefundTransfer(output *Output, userID, followID string, origin ActionType, errCode ErrorCode) (*Transfer, error) {
	transferAction := TransferAction{
		Code:     int(errCode),
		Origin:   origin,
		Source:   ActionTypeRefundTransfer,
		FollowID: followID,
	}
	memoStr, e := transferAction.Format()
	if e != nil {
		return nil, e
	}

	return &Transfer{
		TraceID:   uuidutil.Modify(output.TraceID, "compound_refund"),
		Opponents: []string{userID},
		Threshold: 1,
		AssetID:   output.AssetID,
		Amount:    output.Amount,
		Memo:      memoStr,
	}, nil
}

func NewOutTransfer(userID, followID, outputTraceID, assetID string, amount decimal.Decimal, transferAction *TransferAction) (*Transfer, error) {
	memoStr, e := transferAction.Format()
	if e != nil {
		return nil, e
	}

	modifier := fmt.Sprintf("%s.%d", followID, transferAction.Source)
	return &Transfer{
		TraceID:   uuidutil.Modify(outputTraceID, modifier),
		Opponents: []string{userID},
		Threshold: 1,
		AssetID:   assetID,
		Amount:    amount,
		Memo:      memoStr,
	}, nil
}

func NewTransfer(traceID, assetID string, amount decimal.Decimal, opponent string) (*Transfer, error) {
	return &Transfer{
		TraceID:   traceID,
		AssetID:   assetID,
		Amount:    amount,
		Threshold: 1,
		Opponents: []string{opponent},
	}, nil
}
