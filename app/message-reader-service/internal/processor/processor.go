package processor

import "context"

type Processor[T any] interface {
	Process(ctx context.Context, event *T) error
}
