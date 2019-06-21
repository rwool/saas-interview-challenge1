package queuemock

import (
	"context"
	"sync"
)

// QueueMock is a mock implementation of the queue.Queue type.
//
// Intended for testing only.
type QueueMock struct {
	Data *sync.Map
}

// New returns a new QueueMock.
func New() *QueueMock {
	return &QueueMock{Data: new(sync.Map)}
}

func (q *QueueMock) getChan(key string) chan []byte {
	var v interface{}
	var ok bool
	v, ok = q.Data.Load(key)
	if !ok {
		c := make(chan []byte, 100)
		v, _ = q.Data.LoadOrStore(key, c)
	}
	return v.(chan []byte)
}

// Push pushes Data to the given channel.
func (q *QueueMock) Push(ctx context.Context, channel string, data [][]byte) error {
	c := q.getChan(channel)
	for _, d := range data {
		c <- d
	}
	return nil
}

// Pull pull Data from the given channel.
//
// If the channel has no Data available, then this call will block until there
// is.
func (q *QueueMock) Pull(ctx context.Context, channel string) ([]byte, error) {
	c := q.getChan(channel)
	out := <-c
	return out, nil
}
