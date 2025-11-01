package handler

import "github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"

type Handler interface {
	Handle(msg model.Message) error
}
