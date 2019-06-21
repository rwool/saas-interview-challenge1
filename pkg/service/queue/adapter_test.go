//+build integration

package queue_test

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rwool/saas-interview-challenge1/pkg/service/internal/redistest"

	"github.com/rwool/saas-interview-challenge1/pkg/service/queue"

	"github.com/stretchr/testify/assert"

	"golang.org/x/sync/errgroup"
)

var seedOnce sync.Once

func randString() string {
	seedOnce.Do(func() { rand.Seed(time.Now().UnixNano()) })
	i := rand.Int()
	return strconv.Itoa(i)
}

func TestRedisConnection(t *testing.T) {
	t.Parallel()
	client := redistest.Connect(t)
	assert.NoError(t, client.Ping().Err(), "Should be no error with Redis connection.")
}

func TestQueue(t *testing.T) {
	//t.Parallel()
	client := redistest.Connect(t)
	adapter := queue.NewRedisAdapter(client)

	id := randString()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)

	// Wait on a channel to "burst" all requests as fast as possible.
	var ready sync.WaitGroup
	ready.Add(11)
	start := make(chan struct{})

	found := make(map[string]struct{})
	var foundMu sync.Mutex

	group.Go(func() error {
		ready.Done()
		<-start
		for i := 0; i < 10; i++ {
			err := adapter.Push(ctx, id, [][]byte{
				[]byte(strconv.Itoa(i)),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	for i := 0; i < 10; i++ {
		group.Go(func() error {
			ready.Done()
			<-start
			data, err := adapter.Pull(ctx, id)
			if err != nil {
				return err
			}
			foundMu.Lock()
			found[string(data)] = struct{}{}
			foundMu.Unlock()
			return nil
		})
	}

	ready.Wait()
	close(start)
	err := group.Wait()
	require.NoError(t, err, "Sending and receiving messages should not error.")

	for i := 0; i < 10; i++ {
		iStr := strconv.Itoa(i)
		_, ok := found[iStr]
		require.True(t, ok, "Missing value %s in found set", iStr)
	}
}

// TODO: Add tests for multi-send, tests with mocks for error handling tests,
//  large payloads, empty payloads, etc.
