// Package keyvalue implements support for storing and retrieving key value
// pairs.
package keyvalue

import (
	"context"
	"time"
)

// KeyValue wraps the set of methods for storing and retrieving data identified
// by a given key.
type KeyValue interface {
	Store(ctx context.Context, key string, data []byte, expiration time.Duration) error
	Retrieve(ctx context.Context, key string) ([]byte, error)

	SetCounter(ctx context.Context, key string, value int64) error
	GetCounter(ctx context.Context, key string) (int64, error)
	IncrementCounter(ctx context.Context, key string) error
}
