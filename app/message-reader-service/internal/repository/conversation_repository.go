package repository

import (
	"context"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"
)

type ConversationRepository interface {
	// GetConversation retrieves messages for a given conversation ID and before/after cursors.
	GetConversationMessages(ctx context.Context, conversationID string, before string, after string) ([]*model.Message, error)
}
