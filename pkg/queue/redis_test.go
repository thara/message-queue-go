package queue

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestNewRedisQueue(t *testing.T) {
	config := &Config{
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		RedisDB:           1, // Use different DB for tests
		MaxRetries:        3,
		DefaultVisibility: 30 * time.Second,
	}

	// This test requires Redis to be running
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	client.Close()

	queue, err := NewRedisQueue(config)
	if err != nil {
		t.Fatalf("Failed to create Redis queue: %v", err)
	}
	defer queue.Close()

	if queue.config.RedisAddr != config.RedisAddr {
		t.Errorf("expected redis addr %s, got %s", config.RedisAddr, queue.config.RedisAddr)
	}
}

func TestRedisQueueIntegration(t *testing.T) {
	config := &Config{
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		RedisDB:           1, // Use different DB for tests
		MaxRetries:        3,
		DefaultVisibility: 30 * time.Second,
	}

	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	// Clean up test data
	client.FlushDB(ctx)
	client.Close()

	queue, err := NewRedisQueue(config)
	if err != nil {
		t.Fatalf("Failed to create Redis queue: %v", err)
	}
	defer queue.Close()

	queueName := "test-queue"
	payload := []byte("test message")

	// Test PushMessage
	err = queue.PushMessage(ctx, queueName, payload)
	if err != nil {
		t.Fatalf("Failed to push message: %v", err)
	}

	// Test PullMessage
	msg, err := queue.PullMessage(ctx, queueName, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to pull message: %v", err)
	}

	if string(msg.Payload) != string(payload) {
		t.Errorf("expected payload %s, got %s", string(payload), string(msg.Payload))
	}

	if msg.Receipt == "" {
		t.Error("expected non-empty receipt")
	}

	// Test Ack
	err = queue.Ack(ctx, msg.ID, msg.Receipt)
	if err != nil {
		t.Fatalf("Failed to acknowledge message: %v", err)
	}

	// Test pulling from empty queue
	_, err = queue.PullMessage(ctx, queueName, 1*time.Second)
	if err != ErrQueueEmpty {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}

	if id1 == id2 {
		t.Error("expected unique IDs")
	}

	if len(id1) != 32 { // 16 bytes * 2 hex chars
		t.Errorf("expected ID length 32, got %d", len(id1))
	}
}