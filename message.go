package uviews

import (
	"encoding/json"
	"time"
)

// Message - All Posts & Puts must have the following structure
type Message struct {
	// Timestamp when message was sent
	Timestamp int64 `json:"timestamp"`
	// Entity
	Data json.RawMessage `json:"data,omitempty"`
}

func NewMessageSim(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      data,
	}
}
