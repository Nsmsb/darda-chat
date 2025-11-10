package model

import "encoding/json"

const (
	// Event types
	EventTypeMessage      = "Message"
	EventTypeMessageEvent = "MessageEvent"
)

// Represents a generic event wrapper
type Event struct {
	Type    string          `json:"type"`    // "Message", "MessageEvent"
	Content json.RawMessage `json:"content"` // Raw JSON, to decode later depending on Type
}
