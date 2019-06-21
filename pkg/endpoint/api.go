package endpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/rwool/saas-interview-challenge1/pkg/service"
)

// MessageMetadata contains the request metadata include the trace ID.
// TODO: Implement this.
type MessageMetadata struct{}

// ProcessDocumentResponse contains the response for a call to the
// ProcessDocument endpoint.
type ProcessDocumentResponse struct {
	MessageMetadata
	service.DocumentFrequenciesResponse
	e error
}

// MakeAPIProcessDocumentEndpoint creates an endpoint for processing documents.
func MakeAPIProcessDocumentEndpoint(a service.APIService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		req := request.(service.DocumentRequest)
		dfr, err := a.ProcessDocument(ctx, req)
		return ProcessDocumentResponse{
			MessageMetadata:             MessageMetadata{},
			DocumentFrequenciesResponse: dfr,
			e:                           err,
		}, nil
	}
}
