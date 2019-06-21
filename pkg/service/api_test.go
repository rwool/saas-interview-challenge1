package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/stretchr/testify/require"

	"github.com/rwool/saas-interview-challenge1/pkg/internal/keyvaluemock"
	"github.com/rwool/saas-interview-challenge1/pkg/internal/queuemock"
	"github.com/rwool/saas-interview-challenge1/pkg/service"
)

func TestAPI(t *testing.T) {
	const channel = "worker"
	l := log.NewNopLogger()
	q := queuemock.New()
	kv := keyvaluemock.New()
	apiService := service.NewAPIService(q, kv, channel, l)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		for {
			d, err := q.Pull(ctx, channel)
			require.NoError(t, err, "Pull from queue should succeed.")

			var dfr service.DocumentID
			err = json.Unmarshal(d, &dfr)
			require.NoError(t, err, "Response should unmarshal successfully.")

			err = kv.Store(ctx, dfr.ID, d, 0)
			require.NoError(t, err, "Should store data successfully.")
		}
	}()

	_, err := apiService.ProcessDocument(ctx, service.DocumentRequest{
		Document:        "This is a document",
		DurationSeconds: 0,
	})
	require.NoError(t, err, "Processing document should not error.")
}
