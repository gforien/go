package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		headers    map[string]string
		wantStatus int
		wantHeader map[string]string
		wantBody   []string
	}{
		{
			name:       "GET /",
			method:     http.MethodGet,
			target:     "/",
			body:       "",
			wantStatus: http.StatusOK,
			wantBody:   []string{"GET /"},
		},
		{
			name:   "POST with query, custom header, and body",
			method: http.MethodPost,
			target: "/foo?bar=baz",
			body:   "hello body",
			headers: map[string]string{
				"X-Custom-Header": "abc123",
			},
			wantStatus: http.StatusOK,
			wantHeader: map[string]string{
				"Content-Type": "text/plain",
			},
			wantBody: []string{
				"POST /foo?bar=baz",
				"X-Custom-Header: abc123",
				"hello body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				tt.method, tt.target, bytes.NewBufferString(tt.body),
			)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()
			httpHandler(rr, req)

			res := rr.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					panic(err)
				}
			}()

			// Status check
			if res.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, res.StatusCode)
			}

			// Headers check
			for k, v := range tt.wantHeader {
				if got := res.Header.Get(k); got != v {
					t.Errorf("expected header %s=%q, got %q", k, v, got)
				}
			}

			// Body check
			data, _ := io.ReadAll(res.Body)
			dump := string(data)
			for _, substr := range tt.wantBody {
				if !strings.Contains(dump, substr) {
					t.Errorf("expected body to contain %q, got:\n%q", substr, dump)
				}
			}
		})
	}
}
