package client

import (
	"fmt"

	"github.com/nicolasacchi/clicore/cierrors"
)

// APIError is the transport-independent error envelope for pmx. For HTTP
// failures set StatusCode (and ErrorCode for Postmark's numeric code); for
// client-side guards (the write-safety gate) set Kind and leave StatusCode 0.
type APIError struct {
	StatusCode int
	ErrorCode  int // Postmark's numeric ErrorCode (0 when absent)
	Detail     string
	Hint       string

	// Kind classifies client-side guard errors raised before a request
	// (StatusCode stays 0). Currently: "write_locked" (confirm gate).
	Kind string
}

func (e *APIError) Error() string {
	// Client-side guard errors (no HTTP status) render from Detail/Hint.
	if e.Kind != "" && e.StatusCode == 0 {
		msg := e.Detail
		if msg == "" {
			msg = e.Kind
		}
		if e.Hint != "" {
			return fmt.Sprintf("%s — %s", msg, e.Hint)
		}
		return msg
	}
	if e.Detail != "" {
		if e.ErrorCode != 0 {
			return fmt.Sprintf("%d (ErrorCode %d): %s", e.StatusCode, e.ErrorCode, e.Detail)
		}
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Detail)
	}
	return fmt.Sprintf("API error %d", e.StatusCode)
}

// ExitCode delegates to the fleet-canonical table (auth=2, validation=3,
// not_found=4, rate_limited=5, write_locked=6, async_timeout=7, else 1).
// Postmark adaptation: it returns HTTP 422 (Unprocessable Entity) where the
// fleet table expects 400 for validation, so 422 is mapped to exit 3.
func (e *APIError) ExitCode() int {
	if e.Kind == "" && e.StatusCode == 422 {
		return 3
	}
	return cierrors.ExitCodeFor(e.StatusCode, e.Kind)
}

// hintForError returns an actionable hint for the common Postmark failures.
func hintForError(statusCode, errorCode int) string {
	// ErrorCode 10 = "Bad or missing API token" (Postmark returns it under 401
	// and sometimes 422). Point at the right scope.
	if errorCode == 10 {
		return "check your token: --server-token / --account-token, POSTMARK_SERVER_TOKEN / POSTMARK_ACCOUNT_TOKEN, or 'pmx config doctor'"
	}
	switch statusCode {
	case 401:
		return "authentication failed — set the right token (server vs account scope); run 'pmx config doctor'"
	case 403:
		return "forbidden — this token lacks access to the resource"
	case 404:
		return "not found — check the ID / message stream / recipient"
	case 422:
		return "unprocessable — check required fields and the error message above"
	case 429:
		return "rate limited — Postmark caps request rate; back off and retry"
	}
	return ""
}
