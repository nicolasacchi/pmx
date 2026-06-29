package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(srv *httptest.Server, serverToken, accountToken string) *Client {
	c := New(serverToken, accountToken, false)
	c.baseURL = srv.URL
	c.WithHTTPClient(srv.Client())
	return c
}

func TestServerScopeSetsServerTokenHeader(t *testing.T) {
	var gotServer, gotAccount string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotServer = r.Header.Get("X-Postmark-Server-Token")
		gotAccount = r.Header.Get("X-Postmark-Account-Token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"TotalCount":1,"Messages":[{"MessageID":"abc"}]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv, "srv-token", "acct-token")
	if _, err := c.Get(context.Background(), ScopeServer, "messages/outbound", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotServer != "srv-token" {
		t.Errorf("server token header = %q, want srv-token", gotServer)
	}
	if gotAccount != "" {
		t.Errorf("account token header should be empty on a server-scope call, got %q", gotAccount)
	}
}

func TestAccountScopeSetsAccountTokenHeader(t *testing.T) {
	var gotAccount string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccount = r.Header.Get("X-Postmark-Account-Token")
		_, _ = w.Write([]byte(`{"TotalCount":0,"Servers":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv, "srv-token", "acct-token")
	if _, err := c.Get(context.Background(), ScopeAccount, "servers", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotAccount != "acct-token" {
		t.Errorf("account token header = %q, want acct-token", gotAccount)
	}
}

func TestMissingServerTokenFailsFast(t *testing.T) {
	c := New("", "acct", false) // no server token
	_, err := c.Get(context.Background(), ScopeServer, "server", nil)
	if err == nil {
		t.Fatal("expected error when server token missing")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("want *APIError, got %T", err)
	}
	if apiErr.ExitCode() != 2 {
		t.Errorf("missing-token exit code = %d, want 2", apiErr.ExitCode())
	}
}

func TestParseErrorPostmarkShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"ErrorCode":10,"Message":"Bad or missing API token."}`))
	}))
	defer srv.Close()

	c := newTestClient(srv, "srv", "acct")
	_, err := c.Get(context.Background(), ScopeServer, "server", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("want *APIError, got %T", err)
	}
	if apiErr.ErrorCode != 10 {
		t.Errorf("ErrorCode = %d, want 10", apiErr.ErrorCode)
	}
	// 422 maps to validation (exit 3) in the Postmark adaptation.
	if apiErr.ExitCode() != 3 {
		t.Errorf("422 exit code = %d, want 3", apiErr.ExitCode())
	}
	if apiErr.Hint == "" {
		t.Error("expected a hint for ErrorCode 10")
	}
}

func TestNonIdempotentNotRetriedOnNetworkError(t *testing.T) {
	// Closed server → connection refused. POST must not retry.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := New("srv", "", false)
	c.baseURL = url
	c.maxRetries = 3
	_, err := c.Post(context.Background(), ScopeServer, "email", map[string]string{"From": "a@b.com"})
	if err == nil {
		t.Fatal("expected a network error")
	}
}
