package model

import (
	"encoding/json"
	"time"
)

const (
	// Event types
	EventTypeMessage      = "Message"
	EventTypeMessageEvent = "MessageEvent"
)

// Represents a generic event wrapper
type Event struct {
	Type      string          `json:"type"`      // "Message", "MessageEvent"
	EventID   string          `json:"event_id"`  // Unique ID for the event
	Timestamp time.Time       `json:"timestamp"` // Timestamp when the event was created
	Content   json.RawMessage `json:"content"`   // Raw JSON, to decode later depending on Type
}
