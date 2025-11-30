package service

import (
	"context"
	"fmt"

	pb "github.com/nsmsb/darda-chat/app/message-reader-service/internal/api/message/gen"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/repository"
	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MessageService struct {
	pb.UnimplementedMessageServiceServer
	conversationRepo      repository.ConversationRepository
	conversationCacheRepo repository.ConversationCacheRepository
}

func NewMessageService(conversationRepo repository.ConversationRepository, conversationCacheRepo repository.ConversationCacheRepository) *MessageService {
	return &MessageService{
		conversationRepo:      conversationRepo,
		conversationCacheRepo: conversationCacheRepo,
	}
}

// GetMessages retrieves messages for a given conversation ID.
func (s *MessageService) GetMessages(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	log := logger.FromContext(ctx)

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

	// Getting conversation messages from cache
	convKey := fmt.Sprintf("conversation:%s:%s", conversationID, before+after)
	messages, err := s.conversationCacheRepo.GetConversationMessages(convKey)
	if err != nil {
		return nil, err
	}

	// Load from database if cache miss
	if len(messages) == 0 {
		log.Info("Cache miss for conversation", zap.String("conversationKey", convKey))
		// Getting conversation
		messages, err = s.conversationRepo.GetConversationMessages(ctx, conversationID, before, after)
		if err != nil {
			return nil, err
		}
		// Setting cache
		err = s.conversationCacheRepo.SetConversationMessages(convKey, messages)
		if err != nil {
			return nil, err
		}
		log.Info("Cache hit for conversation", zap.String("conversationKey", convKey), zap.Int("messageCount", len(messages)))
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
