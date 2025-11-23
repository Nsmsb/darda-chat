package model

import "time"

// Represents an actual chat message
type Message struct {
	ID             string    `json:"id" bson:"_id"`
	ConversationID string    `json:"conversation_id" bson:"conversationid"`
	Sender         string    `json:"sender" bson:"sender"`
	Destination    string    `json:"destination" bson:"destination"`
	Content        string    `json:"content" bson:"content"`
	Timestamp      time.Time `json:"timestamp" bson:"timestamp"`
}

type OutboxMessage struct {
	ID          string `json:"id" bson:"_id"`
	Payload     Message
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	ProcessedAt time.Time `json:"processed_at" bson:"processed_at"`
}
