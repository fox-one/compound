package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/jmoiron/sqlx/types"
)

type (
	// Message message struct
	Message struct {
		ID        int64          `sql:"PRIMARY_KEY" json:"id,omitempty"`
		CreatedAt time.Time      `json:"created_at,omitempty"`
		MessageID string         `sql:"size:36" json:"message_id,omitempty"`
		UserID    string         `sql:"size:36" json:"user_id,omitempty"`
		Raw       types.JSONText `sql:"type:TEXT" json:"raw,omitempty"`
	}

	// MessageStore message store interface
	MessageStore interface {
		Create(ctx context.Context, messages []*Message) error
		List(ctx context.Context, limit int) ([]*Message, error)
		Delete(ctx context.Context, messages []*Message) error
	}

	// MessageService message service interface
	MessageService interface {
		Send(ctx context.Context, messages []*Message) error
		Meet(ctx context.Context, userID string) error
	}
)

// BuildMessage build message
func BuildMessage(req *mixin.MessageRequest) *Message {
	raw, _ := json.Marshal(req)
	return &Message{
		MessageID: req.MessageID,
		UserID:    req.RecipientID,
		Raw:       raw,
	}
}
