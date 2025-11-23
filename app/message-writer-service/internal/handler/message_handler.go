package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageHandler struct {
	dbName               string
	collectionName       string
	outboxCollectionName string
	dbClient             *mongo.Client
}

// NewMessageHandler creates a new MessageHandler instance.
func NewMessageHandler(dbName, collectionName string, dbClient *mongo.Client) *MessageHandler {
	return &MessageHandler{
		dbName:               dbName,
		collectionName:       collectionName,
		outboxCollectionName: fmt.Sprintf("%s_outbox", collectionName),
		dbClient:             dbClient,
	}
}

// Handle handles the message event and writes it to the database
func (h *MessageHandler) Handle(ctx context.Context, event model.Event) error {
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
func (h *MessageHandler) insertMessageWithOutbox(ctx context.Context, msg model.Message) error {
	// start a session
	session, err := h.dbClient.StartSession()
	if err != nil {
		return fmt.Errorf("start session error: %w", err)
	}
	defer session.EndSession(ctx)

	// transaction function
	callback := func(sessionCtx mongo.SessionContext) (interface{}, error) {
		// Insert message into messages collection
		messageCollection := h.dbClient.Database(h.dbName).Collection(h.collectionName)
		_, err := messageCollection.InsertOne(sessionCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("insert message error: %w", err)
		}

		// Insert outbox event into outbox collection
		outboxCollection := h.dbClient.Database(h.dbName).Collection(h.outboxCollectionName)
		outboxEvent := model.OutboxMessage{
			ID:          msg.ID,
			Payload:     msg,
			CreatedAt:   time.Now().UTC(),
			ProcessedAt: time.Time{},
		}
		_, err = outboxCollection.InsertOne(sessionCtx, outboxEvent)
		if err != nil {
			return nil, fmt.Errorf("insert outbox event error: %w", err)
		}

		return nil, nil
	}

	// execute transaction
	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	return nil
}
