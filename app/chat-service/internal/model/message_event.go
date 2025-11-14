package model

import "time"

// Represents message-related events (e.g. delivered, read, typing)
type MessageEvent struct {
	ConversationID string    `json:"conversation_id"`
	Timestamp      time.Time `json:"timestamp"`
}
