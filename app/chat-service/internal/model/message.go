package model

import "time"

// Represents an actual chat message
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Sender         string    `json:"sender" binding:"required"`
	Destination    string    `json:"destination" binding:"required"`
	Content        string    `json:"content" binding:"required"`
	Timestamp      time.Time `json:"timestamp"`
}
