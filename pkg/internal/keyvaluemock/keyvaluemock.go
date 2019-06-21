package keyvaluemock

import (
	"context"
	"sync"
	"time"
)

// KeyValueMock is a mock implementation of the keyvalue.KeyValue type.
type KeyValueMock struct {
	channels *sync.Map
}

// New returns a new KeyValueMock.
func New() *KeyValueMock {
	return &KeyValueMock{channels: new(sync.Map)}
}

type counter struct {
	i  int64
	mu sync.Mutex
}

func (k *KeyValueMock) getVal(key string, fn func() interface{}) interface{} {
	var v interface{}
	var ok bool
	v, ok = k.channels.Load(key)
	if !ok {
		val := fn()
		v, _ = k.channels.LoadOrStore(key, val)
	}
	return v
}

func (k *KeyValueMock) getChan(key string) chan []byte {
	val := k.getVal(key, func() interface{} {
		return make(chan []byte, 100)
	})
	return val.(chan []byte)
}

func (k *KeyValueMock) getCounter(key string) *counter {
	val := k.getVal(key, func() interface{} {
		return &counter{}
	})
	return val.(*counter)
}

// Store stores bytes into key.
func (k *KeyValueMock) Store(ctx context.Context, key string, data []byte, expiration time.Duration) error {
	c := k.getChan(key)
	select {
	case c <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Retrieve retrieves the bytes for key.
func (k *KeyValueMock) Retrieve(ctx context.Context, key string) ([]byte, error) {
	c := k.getChan(key)
	select {
	case v := <-c:
		return v, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// SetCounter sets the value of the counter for key.
func (k *KeyValueMock) SetCounter(ctx context.Context, key string, value int64) error {
	c := k.getCounter(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.i = value
	return nil
}

// GetCounter gets the current value of the counter for key.
func (k *KeyValueMock) GetCounter(ctx context.Context, key string) (int64, error) {
	c := k.getCounter(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.i, nil
}

// IncrementCounter increments the value of the counter for key.
func (k *KeyValueMock) IncrementCounter(ctx context.Context, key string) error {
	c := k.getCounter(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.i++
	return nil
}
