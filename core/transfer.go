package core

import (
	"encoding/base64"
	"encoding/json"
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
