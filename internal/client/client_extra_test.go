package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestRetryOn429ThenSuccess(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(srv, "srv", "")
	data, err := c.Get(context.Background(), ScopeServer, "server", nil)
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if hits.Load() != 2 {
		t.Errorf("expected 2 hits (429 then 200), got %d", hits.Load())
	}
	if string(data) != `{"ok":true}` {
		t.Errorf("body = %s", data)
	}
}

func TestServerErrorExhaustsAndFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(srv, "srv", "")
	c.maxRetries = 0 // fail fast, no backoff sleeps in the test
	_, err := c.Get(context.Background(), ScopeServer, "server", nil)
	if err == nil {
		t.Fatal("expected error on 500")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 500 {
		t.Errorf("want 500 APIError, got %v", err)
	}
}

func TestNoContentReturnsEmptyObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(srv, "srv", "")
	data, err := c.Delete(context.Background(), ScopeServer, "templates/1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("204 body = %q, want {}", data)
	}
}

func TestPostSendsBodyAndContentType(t *testing.T) {
	var gotCT string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"MessageID":"x"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv, "srv", "")
	_, err := c.Post(context.Background(), ScopeServer, "email", map[string]any{"From": "a@b.com", "Subject": "hi"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q", gotCT)
	}
	if gotBody["From"] != "a@b.com" || gotBody["Subject"] != "hi" {
		t.Errorf("body round-trip failed: %v", gotBody)
	}
}

func TestBaseURLEnvOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	t.Setenv("POSTMARK_API_BASE_URL", srv.URL+"/")
	c := New("srv", "", false) // no explicit baseURL; should pick up the env
	if _, err := c.Get(context.Background(), ScopeServer, "server", nil); err != nil {
		t.Fatalf("env-base-url request failed: %v", err)
	}
}
