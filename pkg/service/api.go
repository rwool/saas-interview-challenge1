package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/pkg/errors"

	"github.com/rwool/saas-interview-challenge1/pkg/service/keyvalue"
	"github.com/rwool/saas-interview-challenge1/pkg/service/queue"
)

// APIService is the user accessible service.
type APIService interface {
	ProcessDocument(ctx context.Context, request DocumentRequest) (DocumentFrequenciesResponse, error)
}

// DocumentRequest is a request for a document to be processed.
type DocumentRequest struct {
	Document        string `json:"document"`
	DurationSeconds int    `json:"duration_seconds"`
}

// DocumentFrequenciesResponse is the response for processing a document.
type DocumentFrequenciesResponse struct {
	DocumentID  string
	Frequencies []Frequency
}

type apiService struct {
	q              queue.Queue
	kv             keyvalue.KeyValue
	requestChannel string
	l              log.Logger
}

// shortRetrieve approximates a non-blocking get request by blocking less.
func shortRetrieve(ctx context.Context, kv keyvalue.KeyValue, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()

	d, err := kv.Retrieve(ctx, key)
	cause := errors.Cause(err)
	if cause == context.DeadlineExceeded {
		return nil, nil
	}
	return d, errors.WithStack(err)
}

// ProcessDocument processes a document.
func (a *apiService) ProcessDocument(ctx context.Context, request DocumentRequest) (DocumentFrequenciesResponse, error) {
	var dfr DocumentFrequenciesResponse

	id := createStringSHA256(request.Document)
	workerRequest := DocumentID{
		DocumentRequest: request,
		ID:              id,
	}
	_ = a.l.Log("LEVEL", "DEBUG", "MESSAGE", fmt.Sprintf("API request to process document %s", id))

	// Check if result is already cached.
	v, err := shortRetrieve(ctx, a.kv, id)
	if err != nil {
		return dfr, errors.WithStack(err)
	}
	if v != nil {
		_ = a.l.Log("LEVEL", "DEBUG", "MESSAGE", fmt.Sprintf("API cache hit for document %s", id))
		err = json.Unmarshal(v, &dfr)
		return dfr, errors.WithStack(err)
	} else {
		_ = a.l.Log("LEVEL", "DEBUG", "MESSAGE", fmt.Sprintf("API cache miss for document %s", id))
	}

	// Send request to be processed by worker.
	dr, err := json.Marshal(workerRequest)
	if err != nil {
		return dfr, errors.WithStack(err)
	}
	if err := a.q.Push(ctx, a.requestChannel, [][]byte{dr}); err != nil {
		return dfr, errors.Wrap(err, "unable to publish document request")
	}
	_ = a.l.Log("LEVEL", "DEBUG", "MESSAGE", fmt.Sprintf("Pushed document request %s on channel %s", workerRequest.ID, a.requestChannel))

	// Poll for completion of processing.
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	var data []byte
	for {
		select {
		case <-ticker.C:
			data, err = shortRetrieve(ctx, a.kv, id)
			if err != nil {
				return dfr, errors.Wrap(err, "unable to get document report")
			}
			if data != nil {
				goto RETRIEVED
			}
		case <-ctx.Done():
			_ = a.l.Log("LEVEL", "ERROR", "MESSAGE", "Failed to retrieve value with ID")
			return dfr, errors.WithStack(ctx.Err())
		}
	}
RETRIEVED:
	err = json.Unmarshal(data, &dfr)
	return dfr, errors.WithStack(err)
}

func newAPIService(q queue.Queue, kv keyvalue.KeyValue, channel string, l log.Logger) *apiService {
	return &apiService{
		q:              q,
		kv:             kv,
		requestChannel: channel,
		l:              l,
	}
}

// NewAPIService returns an APIService.
func NewAPIService(q queue.Queue, kv keyvalue.KeyValue, channel string, l log.Logger) APIService {
	return newAPIService(q, kv, channel, l)
}
