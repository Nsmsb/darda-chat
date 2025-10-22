package service

type Connection interface {
	StartReading()
	NewSubscriber() <-chan string
	RemoveSubscriber(ch <-chan string) error
	Close() error
}
