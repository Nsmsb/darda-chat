package service

import (
	"context"
	"slices"

	pb "github.com/nsmsb/darda-chat/app/message-reader-service/internal/api/message/gen"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MessageService struct {
	pb.UnimplementedMessageServiceServer
	client *mongo.Client
}

func NewMessageService() *MessageService {
	return &MessageService{
		client: db.Client(),
	}
}

// GetMessages retrieves messages for a given conversation ID.
func (s *MessageService) GetMessages(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	config := config.Get()

	// Getting Conversation ID from request
	conversationID := request.GetConversationId()
	if conversationID == "" {
		return nil, status.Error(codes.InvalidArgument, "conversation_id is required")
	}

	// Getting collection
	col := s.client.Database(config.MongoDBName).Collection(config.MongoCollectionName)

	// MongoDB filter (by conversationId)
	filter := bson.M{
		"conversationid": conversationID,
	}

	// Setting find options: newest first, limit N
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(config.MessagePageSize))

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "mongo find error: %v", err)
	}
	defer cursor.Close(ctx)

	// Temporary slice for raw documents
	var docs []bson.M

	if err := cursor.All(ctx, &docs); err != nil {
		return nil, status.Errorf(codes.Internal, "cursor decode error: %v", err)
	}

	// Convert to protobuf messages
	messages := make([]*pb.Message, 0, len(docs))

	for _, d := range docs {
		msg := &pb.Message{}

		// id
		if id, ok := d["id"].(string); ok {
			msg.Id = id
		}

		// conversationId
		if conv, ok := d["conversationid"].(string); ok {
			msg.ConversationId = conv
		}

		// sender
		if sender, ok := d["sender"].(string); ok {
			msg.Sender = sender
		}

		// destination
		if dest, ok := d["destination"].(string); ok {
			msg.Destination = dest
		}

		// content
		if content, ok := d["content"].(string); ok {
			msg.Content = content
		}

		// timestamp
		if ts, ok := d["timestamp"].(primitive.DateTime); ok {
			msg.Timestamp = timestamppb.New(ts.Time())
		}

		messages = append(messages, msg)
	}

	// Reverse back to oldest → newest (UI-friendly)
	slices.Reverse(messages)

	return &pb.GetMessagesResponse{
		Messages: messages,
	}, nil
}
