//+build integration

package keyvalue_test

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/rwool/saas-interview-challenge1/pkg/service/internal/redistest"
	"github.com/rwool/saas-interview-challenge1/pkg/service/keyvalue"
)

var seedOnce sync.Once

func randString() string {
	seedOnce.Do(func() { rand.Seed(time.Now().UnixNano()) })
	i := rand.Int()
	return strconv.Itoa(i)
}

func TestStoreRetrieve(t *testing.T) {
	t.Parallel()
	c := redistest.Connect(t)
	rc := keyvalue.NewRedisAdapter(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := t.Name() + randString()
	value := []byte("123")
	err := rc.Store(ctx, key, value, 5*time.Second)
	require.NoError(t, err, "Should store value without error.")

	retValue, err := rc.Retrieve(ctx, key)
	require.NoError(t, err, "Should retrieve value without error.")
	assert.Equal(t, value, retValue, "Stored and retrieved values should match.")
}

func TestIncrementAndGet(t *testing.T) {
	t.Parallel()
	c := redistest.Connect(t)
	rc := keyvalue.NewRedisAdapter(c)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := t.Name() + randString()
	var current int64
	incr := func() {
		err := rc.IncrementCounter(ctx, key)
		require.NoError(t, err, "Should increment counter with no error.")
		current++
	}
	get := func() {
		val, err := rc.GetCounter(ctx, key)
		require.NoError(t, err, "Should get counter with no error.")
		require.Equal(t, current, val, "Incremented value should equal retrieved value.")
	}

	incr()
	get()
	incr()
	incr()
	get()
	get()
	incr()
	get()
}

// TODO: Add more tests to check for race conditions, higher load, larger
//  values, larger keys, counter set, etc.
