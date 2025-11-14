package handler

import (
	"context"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
)

// Handler defines the interface for handling messages events.
type Handler interface {
	// Implement the logic to process a message event.
	Handle(ctx context.Context, event model.Event) error
}
