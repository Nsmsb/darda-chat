package processor

import (
	"context"
)

// Processor defines the interface for handling messages events.
type Processor[T any] interface {
	// Implement the logic to process a message event.
	Process(ctx context.Context, event *T) error
}
