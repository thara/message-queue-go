package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thara/message-queue-go/pkg/queue"
)

func main() {
	config := &queue.Config{
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		RedisDB:           0,
		MaxRetries:        3,
		DefaultVisibility: 30 * time.Second,
	}

	q, err := queue.NewRedisQueue(config)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}
	defer q.Close()

	ctx := context.Background()
	queueName := "tasks"

	fmt.Println("=== Message Queue Example ===")

	// Push messages
	fmt.Println("\n1. Pushing messages to queue...")
	messages := []string{
		"Process order #1234",
		"Send email notification",
		"Generate report",
	}

	for _, msg := range messages {
		err := q.PushMessage(ctx, queueName, []byte(msg))
		if err != nil {
			log.Printf("Failed to push message: %v", err)
		} else {
			fmt.Printf("   ✓ Pushed: %s\n", msg)
		}
	}

	// Pull and process messages
	fmt.Println("\n2. Pulling messages from queue...")
	for i := 0; i < len(messages); i++ {
		msg, err := q.PullMessage(ctx, queueName, 10*time.Second)
		if err == queue.ErrQueueEmpty {
			fmt.Println("   Queue is empty")
			break
		}
		if err != nil {
			log.Printf("Failed to pull message: %v", err)
			continue
		}

		fmt.Printf("   ✓ Pulled: %s (ID: %s)\n", string(msg.Payload), msg.ID[:8])

		// Simulate processing
		time.Sleep(100 * time.Millisecond)

		// Acknowledge the message
		err = q.Ack(ctx, msg.ID, msg.Receipt)
		if err != nil {
			log.Printf("Failed to acknowledge message: %v", err)
		} else {
			fmt.Printf("   ✓ Acknowledged: %s\n", msg.ID[:8])
		}
	}

	// Try pulling from empty queue
	fmt.Println("\n3. Attempting to pull from empty queue...")
	_, err = q.PullMessage(ctx, queueName, 5*time.Second)
	if err == queue.ErrQueueEmpty {
		fmt.Println("   ✓ Queue is empty (as expected)")
	}

	fmt.Println("\n=== Example completed ===")
}