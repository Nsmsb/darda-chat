package repository

import (
	"context"
	"slices"
	"time"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MongoConversationRepository struct {
	client *mongo.Client
}

// NewMongoConversationRepository creates a new instance of MongoConversationRepository.
func NewMongoConversationRepository(client *mongo.Client) *MongoConversationRepository {
	return &MongoConversationRepository{
		client: client,
	}
}

// GetConversation retrieves messages for a given conversation ID and before/after cursors.
// When both before and after are empty, it retrieves the latest messages.
func (r *MongoConversationRepository) GetConversationMessages(ctx context.Context, conversationID string, before string, after string) ([]*model.Message, error) {
	config := config.Get()

	// Getting collection
	col := r.client.Database(config.MongoDBName).Collection(config.MongoCollectionName)

	// MongoDB filter (by conversationId)
	filter := bson.M{
		"conversationid": conversationID,
	}

	// Adding cursor conditions
	if before != "" {
		t, err := time.Parse(time.RFC3339, before)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid before timestamp: %v", err)
		}
		beforeTime := primitive.NewDateTimeFromTime(t)
		filter["timestamp"] = bson.M{"$lt": beforeTime}
	}

	// If after is set, add $gt condition
	if after != "" {
		t, err := time.Parse(time.RFC3339, after)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid after timestamp: %v", err)
		}
		afterTime := primitive.NewDateTimeFromTime(t)
		filter["timestamp"] = bson.M{"$gt": afterTime}
	}

	// Setting find options: newest first, limit set to MessagePageSize
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(config.MessagePageSize))

	// Executing find query
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "mongo find error: %v", err)
	}
	defer cursor.Close(ctx)

	// Temporary slice for raw documents
	var messages []*model.Message

	if err := cursor.All(ctx, &messages); err != nil {
		return nil, status.Errorf(codes.Internal, "cursor decode error: %v", err)
	}

	// Reverse back to oldest → newest (UI-friendly)
	slices.Reverse(messages)

	return messages, nil
}
