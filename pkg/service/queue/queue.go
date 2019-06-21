// Package queue implements support for reading and writing to message queues.
package queue

import (
	"context"
)

// Queue wraps the set of methods for reading and writing to a queue.
//
// A major limitation of this interface is that it does not provide a method of
// acknowledging received messages, so if a service dies for whatever reason,
// there is guaranteed data loss. This is due to basing the design around Redis,
// which does not have equivalent pub/sub mechanisms that would typically be
// found in dedicated message queueing software.
// It may be possible to work around this in Redis with a more complicated
// tracking of messages, but that would fall outside the scope of this project.
type Queue interface {
	Push(ctx context.Context, channel string, data [][]byte) error
	Pull(ctx context.Context, channel string) ([]byte, error)
}
