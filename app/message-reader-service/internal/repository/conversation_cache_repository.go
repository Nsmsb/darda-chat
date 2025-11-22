package repository

import "github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"

type ConversationCacheRepository interface {
	// Set a cache of conversation messages
	SetConversationMessages(conversationKey string, messages []*model.Message) error
	// Get cached conversation messages
	GetConversationMessages(conversationKey string) ([]*model.Message, error)
}
