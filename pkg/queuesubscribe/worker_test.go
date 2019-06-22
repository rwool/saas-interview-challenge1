package queuesubscribe_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/sync/semaphore"

	"github.com/go-kit/kit/log"
	"github.com/rwool/saas-interview-challenge1/pkg/internal/queuemock"
	"github.com/rwool/saas-interview-challenge1/pkg/queuesubscribe"
	"github.com/stretchr/testify/require"
)

func TestMakeQueuePubSubHandler(t *testing.T) {
	t.Parallel()

	var (
		count   = 3
		channel = t.Name()
		q       = queuemock.New()
		l       = log.NewNopLogger()

		// Use a weighted semaphore in place of a sync.WaitGroup to not have the
		// test block forever in the event of an error.
		sema = semaphore.NewWeighted(int64(count))

		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	)
	defer cancel()

	f := func(_ context.Context, request interface{}) (response interface{}, err error) {
		defer sema.Release(1)
		return nil, nil
	}

	config := queuesubscribe.Config{
		Endpoint: f,
		Queue:    q,
		Log:      l,
		Channel:  channel,
	}
	handler := queuesubscribe.MakeWorkerHandler(config)
	go handler(ctx)

	var documents [][]byte
	for i := 0; i < count; i++ {
		documents = append(documents, []byte(`{"document": "One two THREE"}`))
	}

	require.NoError(t, sema.Acquire(ctx, 3), "Semaphore acquisition should happen.")
	err := q.Push(ctx, channel, documents)
	assert.NoError(t, err, "Pushing value should not error.")

	require.NoError(t, sema.Acquire(ctx, 3), "Semaphore acquisition should happen.")
}

// TODO: Add more tests for race conditions, invalid JSON, etc.
