package repository

import "github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"

// ConversationCacheRepository defines the interface for conversation cache repository
// TODO: add context to the methods
type ConversationCacheRepository interface {
	// Set a cache of conversation message
	SetConversationMessage(conversationKey string, messages *model.Message) error
	// Set a cache of conversation messages
	SetConversationMessages(conversationKey string, messages []*model.Message) error
	// Get cached conversation messages
	GetConversationMessages(conversationKey string) ([]*model.Message, error)
}
