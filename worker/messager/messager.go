package messenger

import (
	"compound/core"
	"compound/worker"
	"context"
	"errors"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/robfig/cron/v3"
)

// Messager message worker
type Messager struct {
	worker.BaseJob
	messageStore   core.MessageStore
	messageService core.MessageService
}

// New new message worker
func New(location string, messages core.MessageStore, messagez core.MessageService) *Messager {
	messager := Messager{
		messageStore:   messages,
		messageService: messagez,
	}

	l, _ := time.LoadLocation(location)
	messager.Cron = cron.New(cron.WithLocation(l))
	spec := "@every 1s"
	messager.Cron.AddFunc(spec, messager.Run)
	messager.OnWork = func() error {
		return messager.onWork(context.Background())
	}

	return &messager
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
