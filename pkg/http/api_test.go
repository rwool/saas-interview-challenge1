package http_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rwool/saas-interview-challenge1/pkg/http"
)

func TestHTTP(t *testing.T) {
	t.Parallel()

	t.Run("No Error", func(t *testing.T) {
		t.Parallel()
		f := func(_ context.Context, request interface{}) (response interface{}, err error) {
			return nil, nil
		}
		handler := http.NewAPIHTTPHandler(f, nil)
		req := httptest.NewRequest("POST", "http://something.com/document", strings.NewReader(`{"document": "abcd"}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, 200, rec.Code, "Should have 200 status code.")
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		f := func(_ context.Context, request interface{}) (response interface{}, err error) {
			return nil, errors.New("error")
		}
		handler := http.NewAPIHTTPHandler(f, nil)
		req := httptest.NewRequest("POST", "http://something.com/document", strings.NewReader(`{"document": "abcd"}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, "error", rec.Body.String(), "Error value should be in response.")
		assert.Equal(t, 500, rec.Code, "Should have 500 status code.")
	})
}
