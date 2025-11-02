package handler

import (
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/db"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageHandler struct {
	dbName         string
	collectionName string
	dbCliebnt      *mongo.Client
}

// NewMessageHandler creates a new MessageHandler instance.
func NewMessageHandler(dbName, collectionName string) *MessageHandler {
	return &MessageHandler{
		dbName:         dbName,
		collectionName: collectionName,
		dbCliebnt:      db.Client(),
	}
}

// Handle handles the message and writes it to the database
func (h *MessageHandler) Handle(msg model.Message) error {
	collection := h.dbCliebnt.Database(h.dbName).Collection(h.collectionName)
	_, err := collection.InsertOne(nil, msg)
	return err
}
