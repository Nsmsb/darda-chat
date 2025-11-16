package service

import "github.com/nsmsb/darda-chat/app/chat-service/internal/model"

// MessageReader defines the interface for reading messages from the message-reader-service.
type MessageReader interface {
	// GetMessages retrieves messages for a given conversation ID.
	GetMessages(conversationID string) ([]*model.Message, error)
}
