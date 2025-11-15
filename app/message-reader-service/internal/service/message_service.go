package service

import (
	"context"

	pb "github.com/nsmsb/darda-chat/app/message-reader-service/internal/api/message/gen"
)

type MessageService struct {
	pb.UnimplementedMessageServiceServer
}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (s *MessageService) GetMessages(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	return nil, nil
}
