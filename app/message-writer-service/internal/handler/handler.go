package handler

import (
	"context"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
)

// Handler defines the interface for handling messages.
type Handler interface {
	// Implement the logic to process a message.
	Handle(ctx context.Context, msg model.Message) error
}
