package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	gohttp "net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
	"github.com/rwool/saas-interview-challenge1/pkg/service"

	"github.com/go-kit/kit/transport/http"
)

// NewHTTPHandler returns a handler that makes the API service endpoints
// available via HTTP.
func NewAPIHTTPHandler(endpoint endpoint.Endpoint, options map[string][]http.ServerOption) gohttp.Handler {
	if options == nil {
		options = make(map[string][]http.ServerOption)
	}
	m := gohttp.NewServeMux()
	makeAPIProcessDocumentHandler(m, endpoint, options["ProcessDocument"]...)
	return m
}

type errorResponse struct {
	Error string
}

func encodeAPIProcessDocumentResponse(_ context.Context, w gohttp.ResponseWriter, r interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	if v, ok := r.(endpoint.Failer); ok && v.Failed() != nil {
		w.WriteHeader(gohttp.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: v.Failed().Error()})
		return nil
	}
	err := json.NewEncoder(w).Encode(r)
	return errors.WithStack(err)
}

func decodeAPIProcessDocumentRequest(_ context.Context, req *gohttp.Request) (i interface{}, e error) {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	defer func() {
		err := req.Body.Close()
		if e != nil && err != nil {
			e = errors.Wrapf(e, "multiple errors: %s", err)
			return
		}
		if err != nil {
			e = err
		}
	}()
	var dr service.DocumentRequest
	err := decoder.Decode(&dr)
	if err == io.EOF {
		err = nil
	}

	if dr.Document == "" {
		return nil, errors.New("invalid document")
	}
	return dr, errors.WithStack(err)
}

func makeAPIProcessDocumentHandler(m *gohttp.ServeMux, endpoint endpoint.Endpoint, options ...http.ServerOption) {
	handler := http.NewServer(endpoint,
		decodeAPIProcessDocumentRequest,
		encodeAPIProcessDocumentResponse,
		options...)
	hf := func(w gohttp.ResponseWriter, r *gohttp.Request) {
		if r.Method != gohttp.MethodPost {
			w.WriteHeader(gohttp.StatusMethodNotAllowed)
			_, _ = fmt.Fprintf(w, "Invalid request method %s", r.Method)
			return
		}
		handler.ServeHTTP(w, r)
	}
	m.Handle("/document", gohttp.HandlerFunc(hf))
}
