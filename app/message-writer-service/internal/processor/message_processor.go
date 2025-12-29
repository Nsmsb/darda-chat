package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageProcessor struct {
	messageRepository       repository.MessageRepository
	outboxMessageRepository repository.OutboxMessageRepository
	client                  *mongo.Client
}

// NewMessageProcessor creates a new MessageProcessor instance.
func NewMessageProcessor(messageRepository repository.MessageRepository, outboxMessageRepository repository.OutboxMessageRepository, client *mongo.Client) *MessageProcessor {
	return &MessageProcessor{
		messageRepository:       messageRepository,
		outboxMessageRepository: outboxMessageRepository,
		client:                  client,
	}
}

// Process processes a message event and writes it to the database
func (h *MessageProcessor) Process(ctx context.Context, event *model.Event) error {
	if event.Type == model.EventTypeMessage {
		// Adding message to message and outbox collections
		var msg model.Message
		if err := json.Unmarshal(event.Content, &msg); err != nil {
			return fmt.Errorf("unmarshal event content error: %w", err)
		}

		// Insert message with outbox pattern
		return h.insertMessageWithOutbox(ctx, msg)
	}
	return nil
}

// insertMessageWithOutbox inserts a message into the messages collection and creates an outbox event in a transaction.
func (h *MessageProcessor) insertMessageWithOutbox(ctx context.Context, message model.Message) error {
	// Start a session
	session, err := h.client.StartSession()
	if err != nil {
		return fmt.Errorf("start session error: %w", err)
	}
	defer session.EndSession(ctx)

	// Transaction function
	callback := func(sessionCtx mongo.SessionContext) (interface{}, error) {
		// Insert message into messages collection
		err := h.messageRepository.WriteMessage(sessionCtx, message)
		if err != nil {
			return nil, fmt.Errorf("insert message error: %w", err)
		}

		// Insert outbox event into outbox collection
		err = h.outboxMessageRepository.WriteOutboxMessage(sessionCtx, message)
		if err != nil {
			return nil, fmt.Errorf("insert outbox event error: %w", err)
		}

		return nil, nil
	}

	// Execute transaction
	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	return nil
}
