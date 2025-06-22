package queue

import (
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
		Queue:      queue,
		Payload:    payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}
}