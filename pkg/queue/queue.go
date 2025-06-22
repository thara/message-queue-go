package queue

import (
	"context"
	"time"
)

type Queue interface {
	PushMessage(ctx context.Context, queue string, payload []byte) error
	PullMessage(ctx context.Context, queue string, visibilityTimeout time.Duration) (*Message, error)
	Ack(ctx context.Context, messageID string, receipt string) error
	Close() error
}

type Config struct {
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	MaxRetries        int
	DefaultVisibility time.Duration
	DeadLetterQueue   string
}