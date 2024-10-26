package redisqueue

import "encoding/json"

// Message wraps the data to be published to the queue.
// This is needed to include headers for the message to store necessary values, e.g. trace_id.
type Message struct {
	// Content is the message content.
	Content json.RawMessage `json:"content"`
	Headers Headers         `json:"headers"`
}

type Headers struct {
	Trace map[string]string `json:"trace"`
}
