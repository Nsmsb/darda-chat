package service

import "github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"

type Dispatcher interface {
	Dispatch(model.Message) error
	DeclareExchange() error
}
