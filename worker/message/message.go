package message

import (
	"compound/core"
	"compound/worker"
	"context"
	"errors"

	"github.com/fox-one/pkg/logger"
)

// Messager message worker
type Messager struct {
	worker.TickWorker
	messageStore   core.MessageStore
	messageService core.MessageService
}

// New new message worker
func New(messages core.MessageStore, messagez core.MessageService) *Messager {
	messager := Messager{
		messageStore:   messages,
		messageService: messagez,
	}

	return &messager
}

// Run run worker
func (w *Messager) Run(ctx context.Context) error {
	return w.StartTick(ctx, func(ctx context.Context) error {
		return w.onWork(ctx)
	})
}

func (w *Messager) onWork(ctx context.Context) error {
	log := logger.FromContext(ctx)
	const Limit = 300
	const Batch = 70

	messages, err := w.messageStore.List(ctx, Limit)
	if err != nil {
		log.WithError(err).Error("messengers.ListPair")
		return err
	}

	if len(messages) == 0 {
		return errors.New("list messages: EOF")
	}

	filter := make(map[string]bool)
	var idx int

	for _, msg := range messages {
		if filter[msg.UserID] {
			continue
		}

		messages[idx] = msg
		filter[msg.UserID] = true
		idx++

		if idx >= Batch {
			break
		}
	}

	messages = messages[:idx]
	if err := w.messageService.Send(ctx, messages); err != nil {
		log.WithError(err).Error("messagez.Send")
		return err
	}

	if err := w.messageStore.Delete(ctx, messages); err != nil {
		log.WithError(err).Error("messagez.Delete")
		return err
	}

	return nil
}
