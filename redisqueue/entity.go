package redisqueue

import "encoding/json"

// PublishedMessage wraps the data to be published to the queue.
// This is needed to include headers for the message to store necessary values, e.g. trace_id.
type PublishedMessage struct {
	// Content is the message content.
	Content any     `json:"content"`
	Headers Headers `json:"headers"`
}

// Message wraps the data to be subscribed to the queue.
// The Content is in the json.RawMessage format to be unmarshalled to the expected struct.
type Message struct {
	// Content is the message content.
	Content json.RawMessage `json:"content"`
	Headers Headers         `json:"headers"`
}

type Headers struct {
	Trace map[string]string `json:"trace"`
}
