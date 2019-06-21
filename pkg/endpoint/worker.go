package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/rwool/saas-interview-challenge1/pkg/service"
)

// DocumentRequest contains a document to process.
type DocumentRequest struct {
	service.DocumentRequest
}

// DocumentFrequencyReportResponse contains a DocumentFrequencyReport and an
// error to indicate a failure in the business logic.
type DocumentFrequencyReportResponse struct {
	service.DocumentFrequencyReport
	e error
}

// Failed indicates if there was a business logic failure.
func (d DocumentFrequencyReportResponse) Failed() error {
	return d.e
}

// MakeWorkerParseDocumentEndpoint creates a Go kit endpoint for parsing
// documents.
func MakeWorkerParseDocumentEndpoint(w service.WorkerService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(service.DocumentID)
		dfr, err := w.ParseDocument(ctx, req)
		return DocumentFrequencyReportResponse{
			DocumentFrequencyReport: dfr,
			e:                       err,
		}, nil
	}
}
