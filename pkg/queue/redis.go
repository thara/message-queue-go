package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client            *redis.Client
	config            *Config
	visibilitySetKey  string
	messageHashPrefix string
}

func NewRedisQueue(config *Config) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisQueue{
		client:            client,
		config:            config,
		visibilitySetKey:  "mq:visibility",
		messageHashPrefix: "mq:msg:",
	}, nil
}

func (r *RedisQueue) PushMessage(ctx context.Context, queue string, payload []byte) error {
	messageID := generateID()
	message := &Message{
		ID:        messageID,
		Queue:     queue,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, r.messageHashPrefix+messageID, "data", messageData)
	pipe.LPush(ctx, r.queueKey(queue), messageID)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to push message: %w", err)
	}

	return nil
}

func (r *RedisQueue) PullMessage(ctx context.Context, queue string, visibilityTimeout time.Duration) (*Message, error) {
	messageID, err := r.client.LMove(ctx, r.queueKey(queue), r.processingKey(queue), "RIGHT", "LEFT").Result()
	if err == redis.Nil {
		return nil, ErrQueueEmpty
	}
	if err != nil {
		return nil, fmt.Errorf("failed to pull message: %w", err)
	}

	messageData, err := r.client.HGet(ctx, r.messageHashPrefix+messageID, "data").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get message data: %w", err)
	}

	var message Message
	if err := json.Unmarshal([]byte(messageData), &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	receipt := generateID()
	message.Receipt = receipt
	message.VisibilityUntil = time.Now().Add(visibilityTimeout)

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, r.messageHashPrefix+messageID, "receipt", receipt)
	pipe.ZAdd(ctx, r.visibilitySetKey, redis.Z{
		Score:  float64(message.VisibilityUntil.Unix()),
		Member: fmt.Sprintf("%s:%s", queue, messageID),
	})
	
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to set visibility timeout: %w", err)
	}

	return &message, nil
}

func (r *RedisQueue) Ack(ctx context.Context, messageID string, receipt string) error {
	storedReceipt, err := r.client.HGet(ctx, r.messageHashPrefix+messageID, "receipt").Result()
	if err == redis.Nil {
		return ErrMessageNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get receipt: %w", err)
	}

	if storedReceipt != receipt {
		return ErrInvalidReceipt
	}

	messageData, err := r.client.HGet(ctx, r.messageHashPrefix+messageID, "data").Result()
	if err != nil {
		return fmt.Errorf("failed to get message data: %w", err)
	}

	var message Message
	if err := json.Unmarshal([]byte(messageData), &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.LRem(ctx, r.processingKey(message.Queue), 1, messageID)
	pipe.ZRem(ctx, r.visibilitySetKey, fmt.Sprintf("%s:%s", message.Queue, messageID))
	pipe.Del(ctx, r.messageHashPrefix+messageID)
	
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	return nil
}

func (r *RedisQueue) Close() error {
	return r.client.Close()
}

func (r *RedisQueue) queueKey(queue string) string {
	return fmt.Sprintf("mq:queue:%s", queue)
}

func (r *RedisQueue) processingKey(queue string) string {
	return fmt.Sprintf("mq:processing:%s", queue)
}

