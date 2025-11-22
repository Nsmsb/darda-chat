package service

import (
	"context"

	pb "github.com/nsmsb/darda-chat/app/message-reader-service/internal/api/message/gen"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MessageService struct {
	pb.UnimplementedMessageServiceServer
	conversationRepo repository.ConversationRepository
}

func NewMessageService(conversationRepo repository.ConversationRepository) *MessageService {
	return &MessageService{
		conversationRepo: conversationRepo,
	}
}

// GetMessages retrieves messages for a given conversation ID.
func (s *MessageService) GetMessages(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	// Getting Conversation ID an cursor parameters from request
	conversationID := request.GetConversationId()
	if conversationID == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}
	before := request.GetBefore()
	after := request.GetAfter()

	// Handling error when both after and before cursors are set
	if before != "" && after != "" {
		return nil, status.Error(codes.InvalidArgument, "only one of 'before' or 'after' can be set")
	}

	// Getting conversation
	messages, err := s.conversationRepo.GetConversation(ctx, conversationID, before, after)
	if err != nil {
		return nil, err
	}

	// Convert to protobuf messages
	conversation := make([]*pb.Message, 0, len(messages))

	for _, msg := range messages {
		protoMsg := &pb.Message{}

		// id
		protoMsg.Id = msg.ID

		// conversationId
		protoMsg.ConversationId = msg.ConversationID

		// sender
		protoMsg.Sender = msg.Sender

		// destination
		protoMsg.Destination = msg.Sender

		// content
		protoMsg.Content = msg.Content

		// timestamp
		protoMsg.Timestamp = timestamppb.New(msg.Timestamp)

		conversation = append(conversation, protoMsg)
	}

	return &pb.GetMessagesResponse{
		Messages: conversation,
	}, nil
}
