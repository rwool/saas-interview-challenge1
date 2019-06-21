package keyvalue

import (
	"context"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/go-redis/redis"
)

// NewRedisAdapter creates a Redis client that supports storing and retrieving
// key value pairs.
func NewRedisAdapter(c *redis.Client) *RedisAdapter {
	return &RedisAdapter{c: c}
}

// Ensure RedisAdapter implements the KeyValue interface.
var _ KeyValue = (*RedisAdapter)(nil)

// RedisAdapter adapts a Redis client to support the KeyValue interface.
type RedisAdapter struct {
	c *redis.Client
}

// Store stores a key value pair in Redis.
//
// If expiration is set to 0, then the key will never expire.
func (r *RedisAdapter) Store(ctx context.Context, key string, data []byte, expiration time.Duration) error {
	// TODO: Handle message trace from ctx.
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	client := r.c.WithContext(ctx)
	value := base64.StdEncoding.EncodeToString(data)
	err := client.Set(key, value, expiration).Err()
	return errors.Wrap(err, "error storing key value pair in Redis")
}

// Retrieve retrieves a value for a given key from Redis.
func (r *RedisAdapter) Retrieve(ctx context.Context, key string) ([]byte, error) {
	// TODO: Handle message trace from ctx.
	if len(key) == 0 {
		return nil, errors.New("invalid key")
	}
	client := r.c.WithContext(ctx)
	v, err := client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "unable to retrieve value for key %q from Redis", key)
	}
	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decode value for key %q as base64", key)
	}
	return data, nil
}

// SetCounter sets the counter with the given key to value.
func (r *RedisAdapter) SetCounter(ctx context.Context, key string, value int64) error {
	// TODO: Create child span for trace.
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	client := r.c.WithContext(ctx)
	valString := strconv.FormatInt(value, 10)
	err := client.Set(key, valString, 0).Err()
	return errors.Wrapf(err, "failed to set number for key: %q", key)
}

// GetCounter gets the current value of a counter.
func (r *RedisAdapter) GetCounter(ctx context.Context, key string) (int64, error) {
	// TODO: Create child span for trace.
	if len(key) == 0 {
		return 0, errors.New("invalid key")
	}
	client := r.c.WithContext(ctx)
	current, err := client.Get(key).Result()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get number for key: %q", key)
	}
	v, err := strconv.ParseInt(current, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "unexpected format or not a number for key %q", key)
	}
	return v, nil
}

// IncrementCounter increments the value with the given key.
//
// If the key does not exist, it will be initialized to 0 and incremented.
func (r *RedisAdapter) IncrementCounter(ctx context.Context, key string) error {
	// TODO: Create child span for trace.
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	client := r.c.WithContext(ctx)
	err := client.Incr(key).Err()
	return errors.Wrapf(err, "failed to increment value for key %q", key)
}
