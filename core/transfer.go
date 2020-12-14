package core

import (
	"encoding/base64"
	"encoding/json"
)

// TransferAction  transfer action
type TransferAction struct {
	Source        ActionType `json:"s,omitempty"`
	TransactionID string     `json:"t,omitempty"`
	Message       string     `json:"m,omitempty"`
}

// Format format TransferAction to string
func (t *TransferAction) Format() (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
