package repository

import (
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type OutboxMessageRepository interface {
	WriteOutboxMessage(ctx mongo.SessionContext, message model.Message) error
}
