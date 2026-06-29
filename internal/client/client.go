// Package client is the pmx HTTP layer for the Postmark API. It mirrors the
// fleet (ddx) client — own retry loop, clicore/redact for verbose logging,
// clicore/cierrors-backed exit codes — and adds Postmark's two-scope auth:
// server-token endpoints (sending, messages, bounces, stats, templates,
// streams, suppressions, server, webhooks, inbound rules) and account-token
// endpoints (servers, sender signatures, domains, data removals).
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nicolasacchi/clicore/redact"
)

const (
	MaxRetries = 3
	Timeout    = 30 * time.Second
	BaseURL    = "https://api.postmarkapp.com"
)

// Scope selects which Postmark auth header a request carries.
type Scope int

const (
	// ScopeServer uses X-Postmark-Server-Token (one Postmark "server").
	ScopeServer Scope = iota
	// ScopeAccount uses X-Postmark-Account-Token (account-wide admin).
	ScopeAccount
)

type Client struct {
	http         *http.Client
	serverToken  string
	accountToken string
	baseURL      string
	verbose      bool
	maxRetries   int
}

// New creates a Postmark API client holding both token scopes. A command picks
// the scope per request; the matching token is required only for the scopes it
// actually calls (so a server-token-only setup can still run server commands).
func New(serverToken, accountToken string, verbose bool) *Client {
	return &Client{
		http:         &http.Client{Timeout: Timeout},
		serverToken:  serverToken,
		accountToken: accountToken,
		baseURL:      BaseURL,
		verbose:      verbose,
		maxRetries:   MaxRetries,
	}
}

// WithHTTPClient overrides the transport (used by tests to point at httptest).
func (c *Client) WithHTTPClient(h *http.Client) *Client { c.http = h; return c }

// Get performs a scoped GET and returns the raw JSON body.
func (c *Client) Get(ctx context.Context, scope Scope, path string, params url.Values) (json.RawMessage, error) {
	if err := c.checkToken(scope); err != nil {
		return nil, err
	}
	return c.doRequest(ctx, http.MethodGet, c.buildURL(path, params), nil, scope)
}

// Post performs a scoped POST with an optional JSON body (nil = empty body).
func (c *Client) Post(ctx context.Context, scope Scope, path string, body any) (json.RawMessage, error) {
	return c.write(ctx, http.MethodPost, scope, path, body)
}

// Put performs a scoped PUT with an optional JSON body.
func (c *Client) Put(ctx context.Context, scope Scope, path string, body any) (json.RawMessage, error) {
	return c.write(ctx, http.MethodPut, scope, path, body)
}

// Patch performs a scoped PATCH with an optional JSON body.
func (c *Client) Patch(ctx context.Context, scope Scope, path string, body any) (json.RawMessage, error) {
	return c.write(ctx, http.MethodPatch, scope, path, body)
}

// Delete performs a scoped DELETE and returns Postmark's response body.
func (c *Client) Delete(ctx context.Context, scope Scope, path string) (json.RawMessage, error) {
	if err := c.checkToken(scope); err != nil {
		return nil, err
	}
	return c.doRequest(ctx, http.MethodDelete, c.buildURL(path, nil), nil, scope)
}

func (c *Client) write(ctx context.Context, method string, scope Scope, path string, body any) (json.RawMessage, error) {
	if err := c.checkToken(scope); err != nil {
		return nil, err
	}
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}
	return c.doRequest(ctx, method, c.buildURL(path, nil), bodyBytes, scope)
}

func (c *Client) checkToken(scope Scope) error {
	switch scope {
	case ScopeServer:
		if c.serverToken == "" {
			return &APIError{StatusCode: 401, ErrorCode: 10, Detail: "server token required for this command",
				Hint: "set --server-token, POSTMARK_SERVER_TOKEN, or run 'pmx config add'"}
		}
	case ScopeAccount:
		if c.accountToken == "" {
			return &APIError{StatusCode: 401, ErrorCode: 10, Detail: "account token required for this command",
				Hint: "set --account-token or POSTMARK_ACCOUNT_TOKEN (account-wide admin scope)"}
		}
	}
	return nil
}

func (c *Client) buildURL(path string, params url.Values) string {
	if strings.HasPrefix(path, "http") {
		return path
	}
	u := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u
}

func (c *Client) setHeaders(req *http.Request, scope Scope, hasBody bool) {
	switch scope {
	case ScopeServer:
		req.Header.Set("X-Postmark-Server-Token", c.serverToken)
	case ScopeAccount:
		req.Header.Set("X-Postmark-Account-Token", c.accountToken)
	}
	req.Header.Set("Accept", "application/json")
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
}

func (c *Client) doRequest(ctx context.Context, method, rawURL string, bodyBytes []byte, scope Scope) (json.RawMessage, error) {
	if c.verbose {
		fmt.Fprintf(os.Stderr, "> %s %s\n", method, redact.URL(rawURL))
		if len(bodyBytes) > 0 {
			fmt.Fprintf(os.Stderr, "> Body: %s\n", redact.Body(bodyBytes))
		}
	}

	start := time.Now()
	resp, err := c.doWithRetry(ctx, method, rawURL, bodyBytes, scope)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "< %d %s (%s, %s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), elapsed.Round(time.Millisecond), humanBytes(len(respBody)))
	}

	if resp.StatusCode == http.StatusNoContent {
		return json.RawMessage("{}"), nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if c.verbose {
			fmt.Fprintf(os.Stderr, "< Body: %s\n", redact.Body(respBody))
		}
		return nil, c.parseError(respBody, resp.StatusCode)
	}

	if len(respBody) == 0 {
		return json.RawMessage("{}"), nil
	}
	return json.RawMessage(respBody), nil
}

func (c *Client) doWithRetry(ctx context.Context, method, rawURL string, bodyBytes []byte, scope Scope) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var bodyReader io.Reader
		if len(bodyBytes) > 0 {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		c.setHeaders(req, scope, len(bodyBytes) > 0)

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			// Never retry a non-idempotent verb on a network error — the write
			// may already have committed server-side.
			if !idempotent(method) || attempt == c.maxRetries {
				return nil, lastErr
			}
			delay := retryDelay(attempt, "")
			if c.verbose {
				fmt.Fprintf(os.Stderr, "! request error, retrying in %s (attempt %d/%d)\n", delay, attempt+1, c.maxRetries)
			}
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// A clean 429/5xx reached the server before any state change — safe to
		// retry for any method.
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			retryAfter := resp.Header.Get("Retry-After")
			resp.Body.Close()
			if attempt == c.maxRetries {
				return nil, &APIError{StatusCode: resp.StatusCode, Detail: fmt.Sprintf("failed after %d retries", c.maxRetries)}
			}
			delay := retryDelay(attempt, retryAfter)
			if c.verbose {
				fmt.Fprintf(os.Stderr, "! %d %s, retrying in %s (attempt %d/%d)\n", resp.StatusCode, http.StatusText(resp.StatusCode), delay, attempt+1, c.maxRetries)
			}
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return resp, nil
	}
	return nil, lastErr
}

func (c *Client) parseError(body []byte, statusCode int) *APIError {
	apiErr := &APIError{StatusCode: statusCode}
	var er struct {
		ErrorCode int    `json:"ErrorCode"`
		Message   string `json:"Message"`
	}
	if json.Unmarshal(body, &er) == nil {
		apiErr.ErrorCode = er.ErrorCode
		apiErr.Detail = er.Message
	}
	if apiErr.Detail == "" {
		switch statusCode {
		case 401:
			apiErr.Detail = "authentication failed — bad or missing API token"
		case 403:
			apiErr.Detail = "forbidden"
		case 404:
			apiErr.Detail = "resource not found"
		case 422:
			apiErr.Detail = "unprocessable entity"
		default:
			apiErr.Detail = http.StatusText(statusCode)
		}
	}
	apiErr.Hint = hintForError(statusCode, apiErr.ErrorCode)
	return apiErr
}

func idempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default: // POST, PATCH
		return false
	}
}

func retryDelay(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	base := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	jitter := time.Duration(rand.IntN(500)) * time.Millisecond
	return base + jitter
}

func humanBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1fKB", kb)
	}
	return fmt.Sprintf("%.1fMB", kb/1024)
}
