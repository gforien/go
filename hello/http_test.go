package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFetch(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK response")); err != nil {
				panic(err)
			}
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	})

	listener, err := net.Listen("tcp", ":0") // random port
	if err != nil {
		t.Skip("dual-stack not supported on this system")
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	defer server.Close()

	_, port, _ := net.SplitHostPort(server.Listener.Addr().String())

	mockResolver := func(o *fetchOpts) {
		o.resolve = func(host string) string {
			if host == "mytest.local" {
				return "127.0.0.1"
			}
			return host
		}
	}

	tests := []struct {
		name           string
		url            *url.URL
		opts           []func(*fetchOpts)
		expectedRes    string
		expectedStatus int
	}{
		{
			name:           "GET localhost /",
			url:            &url.URL{Scheme: "http", Host: "localhost:" + port},
			expectedRes:    "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET localhost /ok",
			url:            &url.URL{Scheme: "http", Host: "localhost:" + port, Path: "/ok"},
			expectedRes:    "OK response",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET 127.0.0.1 /",
			url:            &url.URL{Scheme: "http", Host: "127.0.0.1:" + port},
			expectedRes:    "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET 127.0.0.1 /ok",
			url:            &url.URL{Scheme: "http", Host: "127.0.0.1:" + port, Path: "/ok"},
			expectedRes:    "OK response",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET IPv6 /",
			url:            &url.URL{Scheme: "http", Host: "[::1]:" + port},
			expectedRes:    "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET IPv6 /ok",
			url:            &url.URL{Scheme: "http", Host: "[::1]:" + port, Path: "/ok"},
			expectedRes:    "OK response",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET mytest.local /",
			url:            &url.URL{Scheme: "http", Host: "mytest.local:" + port},
			opts:           []func(*fetchOpts){mockResolver},
			expectedRes:    "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET mytest.local /ok",
			url:            &url.URL{Scheme: "http", Host: "mytest.local:" + port, Path: "/ok"},
			opts:           []func(*fetchOpts){mockResolver},
			expectedRes:    "OK response",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE: fetch would panic on error and fail the test
			res, body := fetch(tt.url, tt.opts...)
			if res == nil {
				t.Errorf("expected a response, got %v", res)
			}
			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
			if string(body) != tt.expectedRes {
				t.Errorf("expected body %q, got %q", tt.expectedRes, string(body))
			}
		})
	}
}

func TestHttpHandler(t *testing.T) {
	// Prepare a fake request
	body := "hello body"
	req := httptest.NewRequest(http.MethodPost, "/foo?bar=baz", bytes.NewBufferString(body))
	req.Header.Set("X-Custom-Header", "abc123")

	// Record the response
	rr := httptest.NewRecorder()

	// Call the handler
	httpHandler(rr, req)

	// Extract the response
	resp := rr.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	// Status code check
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Header check
	if got := resp.Header.Get("Content-Type"); got != "text/plain" {
		t.Errorf("expected Content-Type text/plain, got %q", got)
	}

	// Body check
	data, _ := io.ReadAll(resp.Body)
	dump := string(data)

	// The dump should include method, URL, header, and body
	if !strings.Contains(dump, "POST /foo?bar=baz") {
		t.Errorf("expected request line in dump, got:\n%q", dump)
	}
	if !strings.Contains(dump, "X-Custom-Header: abc123") {
		t.Errorf("expected header in dump, got:\n%q", dump)
	}
	if !strings.Contains(dump, body) {
		t.Errorf("expected body in dump, got:\n%q", dump)
	}
}
