package processor

import (
	"context"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
)

// Processor defines the interface for handling messages events.
type Processor interface {
	// Implement the logic to process a message event.
	Process(ctx context.Context, event model.Event) error
}
