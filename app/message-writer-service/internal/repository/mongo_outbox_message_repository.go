package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoOutboxMessageRepository struct {
	client         *mongo.Client
	dbName         string
	collectionName string
}

func NewMongoOutboxMessageRepository(client *mongo.Client, dbName string, collectionName string) *MongoOutboxMessageRepository {
	return &MongoOutboxMessageRepository{
		client:         client,
		dbName:         dbName,
		collectionName: collectionName,
	}
}

func (r *MongoOutboxMessageRepository) Client() *mongo.Client {
	return r.client
}

func (r *MongoOutboxMessageRepository) WriteOutboxMessage(ctx mongo.SessionContext, message model.Message) error {
	collection := r.client.Database(r.dbName).Collection(r.collectionName)
	outboxMessage := model.OutboxMessage{
		ID:          message.ID,
		Payload:     message,
		CreatedAt:   time.Now().UTC(),
		ProcessedAt: time.Time{},
	}
	_, err := collection.InsertOne(ctx, outboxMessage)
	if err != nil {
		return fmt.Errorf("insert outbox event error: %w", err)
	}
	return nil
}

func (r *MongoOutboxMessageRepository) MarkMessageAsProcessed(ctx mongo.SessionContext, message model.OutboxMessage) error {
	collection := r.client.Database(r.dbName).Collection(r.collectionName)

	filter := bson.M{
		"_id": message.ID,
	}

	update := bson.M{
		"$set": bson.M{
			"processedAt": time.Now().UTC(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("mark outbox message as processed error: %w", err)
	}
	return nil
}

func (r *MongoOutboxMessageRepository) GetUnprocessedMessages(ctx mongo.SessionContext, limit int) ([]model.OutboxMessage, error) {
	// TODO: switch to MongoDB Change Streams for real-time processing
	collection := r.client.Database(r.dbName).Collection(r.collectionName)
	filter := map[string]interface{}{
		"processedAt": time.Time{},
	}

	options := options.Find().
		SetLimit(int64(limit)).
		SetSort(map[string]int{"createdAt": 1})

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, fmt.Errorf("find unprocessed outbox messages error: %w", err)
	}
	defer cursor.Close(ctx)

	var outboxMessages []model.OutboxMessage
	for cursor.Next(ctx) {
		var outboxMessage model.OutboxMessage
		if err := cursor.Decode(&outboxMessage); err != nil {
			return nil, fmt.Errorf("decode outbox message error: %w", err)
		}
		outboxMessages = append(outboxMessages, outboxMessage)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return outboxMessages, nil
}

func (r *MongoOutboxMessageRepository) StreamUnprocessedMessages(ctx context.Context) (<-chan model.OutboxMessage, error) {

	collection := r.client.
		Database(r.dbName).
		Collection(r.collectionName)

	// Pipeline: capture only inserts where processedAt is empty
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: "insert"},
			{Key: "fullDocument.processedAt", Value: time.Time{}},
		}}},
	}

	stream, err := collection.Watch(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("watch outbox messages error: %w", err)
	}

	// Output channel
	out := make(chan model.OutboxMessage)

	// Goroutine to continuously stream events
	go func() {
		defer stream.Close(ctx)

		for stream.Next(ctx) {
			var event struct {
				FullDocument model.OutboxMessage `bson:"fullDocument"`
			}

			if err := stream.Decode(&event); err != nil {
				log.Printf("decode change stream error: %v\n", err)
				continue
			}

			// send message to dispatcher
			out <- event.FullDocument
		}

		if err := stream.Err(); err != nil {
			log.Printf("change stream error: %v", err)
		}
	}()

	return out, nil
}
