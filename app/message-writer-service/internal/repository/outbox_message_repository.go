package repository

import (
	"context"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type OutboxMessageRepository interface {
	Client() *mongo.Client
	GetUnprocessedMessages(ctx mongo.SessionContext, limit int) ([]model.OutboxMessage, error)
	StreamUnprocessedMessages(ctx context.Context) (<-chan model.OutboxMessage, error)
	WriteOutboxMessage(ctx mongo.SessionContext, message model.Message) error
	MarkMessageAsProcessed(ctx mongo.SessionContext, message model.OutboxMessage) error
}
