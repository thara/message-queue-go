# Message Queue Go

A Redis-backed message queue implementation in Go providing reliable message delivery with visibility timeout and acknowledgment support.

## Features

- **At-least-once delivery** - Messages are guaranteed to be delivered at least once
- **Visibility timeout** - Messages become invisible to other consumers while being processed
- **Message acknowledgment** - Explicit acknowledgment required to remove messages from queue
- **Simple API** - Just three main methods: PushMessage, PullMessage, and Ack
- **Redis backend** - Leverages Redis for persistence and atomic operations

## Installation

```bash
go get github.com/thara/message-queue-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/thara/message-queue-go/pkg/queue"
)

func main() {
    // Configure the queue
    config := &queue.Config{
        RedisAddr:         "localhost:6379",
        RedisPassword:     "",
        RedisDB:           0,
        DefaultVisibility: 30 * time.Second,
    }
    
    // Create queue instance
    q, err := queue.NewRedisQueue(config)
    if err != nil {
        log.Fatal(err)
    }
    defer q.Close()
    
    ctx := context.Background()
    
    // Push a message
    err = q.PushMessage(ctx, "myqueue", []byte("Hello, World!"))
    
    // Pull a message (with 30 second visibility timeout)
    msg, err := q.PullMessage(ctx, "myqueue", 30*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process the message
    log.Printf("Processing: %s", string(msg.Payload))
    
    // Acknowledge the message
    err = q.Ack(ctx, msg.ID, msg.Receipt)
}
```

## API Reference

### Queue Interface

```go
type Queue interface {
    // PushMessage adds a message to the specified queue
    PushMessage(ctx context.Context, queue string, payload []byte) error
    
    // PullMessage retrieves a message from the queue with visibility timeout
    PullMessage(ctx context.Context, queue string, visibilityTimeout time.Duration) (*Message, error)
    
    // Ack acknowledges a message, removing it from the queue
    Ack(ctx context.Context, messageID string, receipt string) error
    
    // Close closes the queue connection
    Close() error
}
```

### Message Structure

```go
type Message struct {
    ID              string        // Unique message identifier
    Queue           string        // Queue name
    Payload         []byte        // Message content
    Timestamp       time.Time     // Creation timestamp
    RetryCount      int          // Number of retries
    VisibilityUntil time.Time    // When message becomes visible again
    Receipt         string        // Receipt for acknowledgment
}
```

## Development

### Prerequisites

- Go 1.22 or higher
- Redis 6.0 or higher
- Docker and Docker Compose (for local development)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/thara/message-queue-go.git
cd message-queue-go
```

2. Start Redis:
```bash
make docker-up
```

3. Run the example:
```bash
make run-example
```

### Available Commands

```bash
make build          # Build the example binary
make test           # Run tests
make test-coverage  # Run tests with coverage report
make lint           # Run linter
make fmt            # Format code
make docker-up      # Start Redis container
make docker-down    # Stop Redis container
```

## Architecture

The queue uses Redis data structures to ensure reliable message delivery:

- **Lists** - Main queue storage using LPUSH/RPOPLPUSH for FIFO ordering
- **Hashes** - Store message metadata and content
- **Sorted Sets** - Track message visibility timeouts
- **Processing Lists** - Hold messages being processed

### Message Flow

1. **Push**: Message is added to queue list and metadata stored in hash
2. **Pull**: Message atomically moved from queue to processing list
3. **Visibility**: Message tracked in sorted set with expiration timestamp
4. **Ack**: Message removed from processing list and all metadata cleaned up

## Error Handling

The queue provides specific error types for common scenarios:

- `ErrQueueEmpty` - No messages available in the queue
- `ErrMessageNotFound` - Message ID not found during acknowledgment
- `ErrInvalidReceipt` - Receipt doesn't match for acknowledgment

## License

MIT License - see LICENSE file for details