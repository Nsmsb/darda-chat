package service

import "context"

type MessageService interface {
	SendMessage(ctx context.Context, destination string, msg string) error
	SubscribeToMessages(ctx context.Context, channel string) (<-chan string, error)
	UnsubscribeFromMessages(channel string, msgCh <-chan string) error
	Close() error
}
