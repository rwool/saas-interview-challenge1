package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/rwool/saas-interview-challenge1/pkg/internal/keyvaluemock"
	"github.com/rwool/saas-interview-challenge1/pkg/internal/queuemock"

	"github.com/rwool/saas-interview-challenge1/pkg/service"
)

func TestWorker(t *testing.T) {
	q := queuemock.New()
	kv := keyvaluemock.New()
	worker := service.NewWorkerService(service.WorkerServiceConfig{
		Queue:   q,
		KeyVal:  kv,
		Log:     log.NewNopLogger(),
		Channel: "worker_parse_document",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dfr, err := worker.ParseDocument(ctx, service.DocumentID{
		DocumentRequest: service.DocumentRequest{
			Document:        "ABC 123 ABC",
			DurationSeconds: 0,
		},
		ID: "1",
	})
	require.NoError(t, err, "Document parsing should succeed.")
	assert.Equal(t, "1", dfr.DocumentID, "Document IDs should match.")
	require.Len(t, dfr.Frequencies, 2, "Should have two unique words.")
	assert.Equal(t, "ABC", dfr.Frequencies[0].Word, "ABC should be first word.")
	assert.Equal(t, 2, dfr.Frequencies[0].Frequency, "ABC have two occurrences.")
	assert.Equal(t, "123", dfr.Frequencies[1].Word, "123 should be second word.")
	assert.Equal(t, 1, dfr.Frequencies[1].Frequency, "123 have one occurrence.")
}

// TODO: Add tests with more elements, parallel calls, etc.
