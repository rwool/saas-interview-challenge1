package queue

import (
	"context"
	"encoding/base64"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// Ensure RedisAdapter implements Queue.
var _ Queue = (*RedisAdapter)(nil)

// NewRedisAdapter creates a new RedisAdapter.
func NewRedisAdapter(c *redis.Client) *RedisAdapter {
	if c == nil {
		panic("nil queue client")
	}
	return &RedisAdapter{
		c:      c,
		unread: make(map[string][][]byte),
	}
}

// RedisAdapter for a Redis client to implement the Queue interface.
type RedisAdapter struct {
	c      *redis.Client
	unread map[string][][]byte
}

func bytesToString(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func stringToBytes(s string) ([]byte, error) {
	//fmt.Println("STRING S IS...", s)
	b, err := base64.StdEncoding.DecodeString(s)
	//fmt.Printf("STRING S DECODED IS...%s -> %s\n", s, b)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode string into bytes")
	}
	return b, nil
}

// stringsToMultiBytes attempts to base64 decode all the strings in s.
//
// Invalid strings will be dropped.
//
// Only the first error encountered will be returned, if there is one.
func stringsToMultiBytes(s []string) ([][]byte, error) {
	// TODO: Handle invalid strings somehow instead of dropping them.
	out := make([][]byte, 0, len(s))
	var err error
	for _, v := range s {
		b, e := stringToBytes(v)
		if e != nil {
			if err == nil {
				err = errors.WithStack(e)
			}
		}
		out = append(out, b)
	}
	return out, err
}

func bytesToInterfaces(b [][]byte) []interface{} {
	out := make([]interface{}, len(b))
	for i, v := range b {
		out[i] = bytesToString(v)
	}
	return out
}

// Push pushes a number of messages to a queue queue.
//
// This function function is NOT thread-safe.
func (r *RedisAdapter) Push(ctx context.Context, channel string, data [][]byte) error {
	// TODO: Handle message trace from ctx.
	if len(data) == 0 {
		return nil
	}
	client := r.c.WithContext(ctx)
	messages := bytesToInterfaces(data)
	err := client.RPush(channel, messages...).Err()
	return errors.Wrapf(err, "error pushing to Redis list \"%s\"", channel)
}

// Pull pulls values from the queue in Redis.
//
// If there is an error, nil will be returned for the bytes. Subsequent calls
// will return the retrieved values, if there were any.
//
// This function is thread-safe.
func (r *RedisAdapter) Pull(ctx context.Context, channel string) ([]byte, error) {
	// TODO: Handle message trace from ctx.
	getUnread := func() []byte {
		unread := r.unread[channel]
		if len(unread) > 0 {
			out := unread[0]
			r.unread[channel] = unread[1:]
			return out
		}
		return nil
	}
	if tmp := getUnread(); tmp != nil {
		return tmp, nil
	}

	client := r.c.WithContext(ctx)
	// Block forever since context cancellation can be used.
	values, err := client.BLPop(0, channel).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from Redis list \"%s\"", channel)
	}

	// Ignore the channel name.
	if len(values) >= 1 {
		values = values[1:]
	}

	b, err := stringsToMultiBytes(values)
	r.unread[channel] = append(r.unread[channel], b...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return getUnread(), nil
}
