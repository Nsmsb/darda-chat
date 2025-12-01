package repository

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"
	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MongoConversationRepository struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	pageSize       int
	collection     *mongo.Collection
}

// NewMongoConversationRepository creates a new instance of MongoConversationRepository.
func NewMongoConversationRepository(client *mongo.Client, dbName string, collectionName string, pageSize int) *MongoConversationRepository {
	return &MongoConversationRepository{
		client:         client,
		dbName:         dbName,
		collectionName: collectionName,
		pageSize:       pageSize,
		collection:     client.Database(dbName).Collection(collectionName),
	}
}

// GetConversation retrieves messages for a given conversation ID and before/after cursors.
// When both before and after are empty, it retrieves the latest messages.
func (r *MongoConversationRepository) GetConversationMessages(ctx context.Context, conversationID string, before string, after string) ([]*model.Message, error) {
	log := logger.FromContext(ctx)
	log.Info("Fetching conversation messages from MongoDB", zap.String("conversationID", conversationID), zap.String("before", before), zap.String("after", after))

	// Handling error when both after and before cursors are set
	if before != "" && after != "" {
		return nil, status.Error(codes.InvalidArgument, "only one of 'before' or 'after' can be set")
	}

	// MongoDB filter (by conversationId)
	filter := bson.M{
		"conversationId": conversationID,
	}

	// Adding cursor conditions
	if before != "" {
		splittedCursor := strings.SplitN(before, "_", 2)
		if len(splittedCursor) != 2 {
			return nil, status.Errorf(codes.InvalidArgument, "invalid before cursor format")
		}

		cursorTs, cursorID := splittedCursor[0], splittedCursor[1]

		t, err := time.Parse(time.RFC3339, cursorTs)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid before timestamp: %v", err)
		}
		beforeTime := primitive.NewDateTimeFromTime(t)

		filter["$or"] = []bson.M{
			{
				"timestamp": bson.M{"$lt": beforeTime},
			},
			{
				"timestamp": beforeTime,
				"_id":       bson.M{"$lt": cursorID},
			},
		}
	}

	// If after is set, add $gt condition
	if after != "" {
		splittedCursor := strings.SplitN(after, "_", 2)
		if len(splittedCursor) != 2 {
			return nil, status.Errorf(codes.InvalidArgument, "invalid after cursor format")
		}

		cursorTs, cursorID := splittedCursor[0], splittedCursor[1]

		t, err := time.Parse(time.RFC3339, cursorTs)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid after timestamp: %v", err)
		}

		afterTime := primitive.NewDateTimeFromTime(t)

		filter["$or"] = []bson.M{
			{
				"timestamp": bson.M{"$gt": afterTime},
			},
			{
				"timestamp": afterTime,
				"_id":       bson.M{"$gt": cursorID},
			},
		}
	}

	// Setting find options: newest first, limit set to pageSize
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(int64(r.pageSize))

	// Executing find query
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "mongo find error: %v", err)
	}
	defer cursor.Close(ctx)

	var messages []*model.Message

	if err := cursor.All(ctx, &messages); err != nil {
		return nil, status.Errorf(codes.Internal, "cursor decode error: %v", err)
	}

	// Reverse back to oldest → newest (UI-friendly)
	slices.Reverse(messages)

	return messages, nil
}
