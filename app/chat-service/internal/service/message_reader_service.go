package service

import (
	"context"

	pb "github.com/nsmsb/darda-chat/app/chat-service/internal/api/message/gen"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/model"
)

type MessageReaderService struct {
	client pb.MessageServiceClient
}

func NewMessageReaderService(client pb.MessageServiceClient) *MessageReaderService {
	return &MessageReaderService{
		client: client,
	}
}

// GetMessages retrieves messages for a given conversation ID using the message-reader-service.
func (s *MessageReaderService) GetMessages(ctx context.Context, conversationID string, before string, after string) ([]*model.Message, error) {
	// Prepare request
	req := &pb.GetMessagesRequest{
		ConversationId: conversationID,
		Before:         before,
		After:          after,
	}
	// Call gRPC method
	resp, err := s.client.GetMessages(ctx, req)
	if err != nil {
		return nil, err
	}
	// Transform response to model.Message slice, initializing empty slice to return [] instead of null when no result is found
	messages := []*model.Message{}
	for _, msg := range resp.Messages {
		messages = append(messages, &model.Message{
			ID:             msg.GetId(),
			ConversationID: msg.GetConversationId(),
			Sender:         msg.GetSender(),
			Destination:    msg.GetDestination(),
			Content:        msg.GetContent(),
			Timestamp:      msg.GetTimestamp().AsTime().UTC(),
		})
	}

	return messages, nil
}
