package queue

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Message struct {
	ID              string
	Queue           string
	Payload         []byte
	Timestamp       time.Time
	RetryCount      int
	VisibilityUntil time.Time
	Receipt         string
}

type MessageOptions struct {
	VisibilityTimeout time.Duration
	MaxRetries        int
}

func NewMessage(queue string, payload []byte) *Message {
	return &Message{
		ID:         generateID(),
		Queue:      queue,
		Payload:    payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}