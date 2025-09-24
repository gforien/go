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
