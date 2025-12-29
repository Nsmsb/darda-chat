package repository

import (
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepository interface {
	Client() *mongo.Client
	WriteMessage(ctx mongo.SessionContext, message model.Message) error
}
