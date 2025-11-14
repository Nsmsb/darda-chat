package service

import "context"

type Publisher interface {
	Publish(ctx context.Context, msg string, queue string) error
	Close() error
}
