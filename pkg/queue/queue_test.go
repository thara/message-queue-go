package queue

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	payload := []byte("test message")
	msg := NewMessage("test-queue", payload)

	if msg.Queue != "test-queue" {
		t.Errorf("expected queue 'test-queue', got %s", msg.Queue)
	}

	if string(msg.Payload) != "test message" {
		t.Errorf("expected payload 'test message', got %s", string(msg.Payload))
	}

	if msg.RetryCount != 0 {
		t.Errorf("expected retry count 0, got %d", msg.RetryCount)
	}

	if msg.ID == "" {
		t.Error("expected non-empty message ID")
	}

	if msg.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestConfig(t *testing.T) {
	config := &Config{
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		RedisDB:           0,
		MaxRetries:        3,
		DefaultVisibility: 30 * time.Second,
		DeadLetterQueue:   "dlq",
	}

	if config.RedisAddr != "localhost:6379" {
		t.Errorf("expected redis addr 'localhost:6379', got %s", config.RedisAddr)
	}

	if config.MaxRetries != 3 {
		t.Errorf("expected max retries 3, got %d", config.MaxRetries)
	}

	if config.DefaultVisibility != 30*time.Second {
		t.Errorf("expected default visibility 30s, got %v", config.DefaultVisibility)
	}
}

func TestErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrQueueEmpty", ErrQueueEmpty, "queue is empty"},
		{"ErrMessageNotFound", ErrMessageNotFound, "message not found"},
		{"ErrInvalidReceipt", ErrInvalidReceipt, "invalid receipt"},
		{"ErrMaxRetriesReached", ErrMaxRetriesReached, "max retries reached"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.msg {
				t.Errorf("expected error message '%s', got '%s'", tc.msg, tc.err.Error())
			}
		})
	}
}