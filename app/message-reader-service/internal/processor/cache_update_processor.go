package processor

import (
	"context"
	"fmt"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/repository"
)

type CacheUpdateProcessor struct {
	repository repository.ConversationCacheRepository
}

func NewCacheUpdateProcessor(repository repository.ConversationCacheRepository) *CacheUpdateProcessor {
	return &CacheUpdateProcessor{
		repository: repository,
	}
}

func (p *CacheUpdateProcessor) Process(ctx context.Context, event *model.Message) error {
	convKey := fmt.Sprintf("conversation:%s:", event.ConversationID)
	return p.repository.SetConversationMessage(convKey, event)
}
