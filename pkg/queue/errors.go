package queue

import "errors"

var (
	ErrQueueEmpty      = errors.New("queue is empty")
	ErrMessageNotFound = errors.New("message not found")
	ErrInvalidReceipt  = errors.New("invalid receipt")
	ErrMaxRetriesReached = errors.New("max retries reached")
)