# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Testing
```bash
make build          # Build example binary to bin/example
make test           # Run all tests with race detection
make test-coverage  # Generate test coverage report (coverage.html)
make lint           # Run golangci-lint (installs if missing)
make fmt            # Format code and tidy modules
```

### Local Development
```bash
make docker-up      # Start Redis container on port 6379
make docker-down    # Stop Redis container
make run-example    # Build and run example (requires Redis)
```

### Dependencies
```bash
make deps           # Download and verify dependencies
```

## Architecture Overview

This is a Redis-backed message queue implementation providing reliable message delivery with at-least-once guarantees.

### Core Components

**Queue Interface** (`pkg/queue/queue.go`)
- `PushMessage(ctx, queue, payload)` - Add message to queue
- `PullMessage(ctx, queue, visibilityTimeout)` - Retrieve and lock message
- `Ack(ctx, messageID, receipt)` - Acknowledge and remove message

**Redis Implementation** (`pkg/queue/redis.go`)
- Uses Redis Lists for FIFO queue storage
- Uses Redis Hashes for message metadata storage  
- Uses Redis Sorted Sets for visibility timeout tracking
- Atomic operations via pipelining for consistency

### Redis Data Structures

**Queue Storage:**
- `mq:queue:{name}` - List containing message IDs in FIFO order
- `mq:processing:{name}` - List holding messages being processed

**Message Storage:**
- `mq:msg:{messageID}` - Hash containing message data and receipt
- `mq:visibility` - Sorted set tracking visibility timeouts

### Message Flow

1. **Push**: LPUSH to queue list + HSET message data
2. **Pull**: RPOPLPUSH queue→processing + set visibility timeout + generate receipt
3. **Ack**: Verify receipt + LREM from processing + ZREM from visibility + DEL message

### Key Design Decisions

- **Visibility Timeout**: Messages become invisible to other consumers while being processed
- **Receipt System**: Each pulled message gets unique receipt for safe acknowledgment
- **At-least-once Delivery**: Messages remain in processing list until explicitly acknowledged
- **FIFO Ordering**: Uses Redis Lists with RPOPLPUSH for atomic dequeue operations

### Testing Approach

- Unit tests should mock Redis operations
- Integration tests require running Redis instance
- Use `make docker-up` to start Redis for testing
- Race detection enabled by default in test runs

### Error Handling

Custom error types defined in `pkg/queue/errors.go`:
- `ErrQueueEmpty` - No messages available
- `ErrMessageNotFound` - Invalid message ID during ack
- `ErrInvalidReceipt` - Receipt mismatch during ack

### Configuration

All Redis connection and queue behavior configured via `queue.Config`:
- Redis connection details (addr, password, db)
- Default visibility timeout
- Max retries and dead letter queue settings