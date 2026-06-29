package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// capturedReq records what HTTP request(s) a command produced, so E2E tests
// assert on the request (method/path/query/scope/body) without capturing stdout.
// Writes are mutex-guarded because composite commands (otp, overview, config
// doctor) fan out concurrent requests — Paths holds every request path.
type capturedReq struct {
	mu           sync.Mutex
	Hit          bool
	Method       string
	Path         string
	RawQuery     string
	ServerToken  string
	AccountToken string
	Body         []byte
	Paths        []string
}

// tmpJSON writes content to a temp file and returns its path, for `--file`
// body tests (the harness feeds no stdin, so `--file -` would read empty input).
func tmpJSON(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp body: %v", err)
	}
	return path
}

// hitPath reports whether any request targeted p. Safe to call after the server
// is closed (httptest.Close waits for in-flight handlers to finish).
func (c *capturedReq) hitPath(p string) bool {
	for _, x := range c.Paths {
		if x == p {
			return true
		}
	}
	return false
}

// resetFlags restores every flag in the command tree to its default + Changed
// false, so the package-level flag vars don't leak between E2E runs (cobra does
// not reset them on its own).
func resetFlags(c *cobra.Command) {
	reset := func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	}
	c.Flags().VisitAll(reset)
	c.PersistentFlags().VisitAll(reset)
	for _, sub := range c.Commands() {
		resetFlags(sub)
	}
}

// silenceStdout sends os.Stdout (where the output package writes) to /dev/null
// for the duration of a command run, keeping `go test` output clean.
func silenceStdout() func() {
	old := os.Stdout
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return func() {}
	}
	os.Stdout = devnull
	return func() { os.Stdout = old; _ = devnull.Close() }
}

// runCmd drives the real root command against an httptest server and captures
// the request(s) it produced. status/respBody define the canned response. Both
// token scopes are set in the env, so any command runs; assert on
// cr.ServerToken / cr.AccountToken to verify the scope used. For single-request
// commands cr.Path/Method/RawQuery are deterministic; for composite commands use
// cr.hitPath(...). Tests must NOT call t.Parallel() — runCmd swaps os.Stdout and
// uses t.Setenv.
func runCmd(t *testing.T, status int, respBody string, args ...string) (*capturedReq, error) {
	t.Helper()
	if status == 0 {
		status = 200
	}
	if respBody == "" {
		respBody = "{}"
	}
	cr := &capturedReq{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		cr.mu.Lock()
		cr.Hit = true
		cr.Method = r.Method
		cr.Path = r.URL.Path
		cr.RawQuery = r.URL.RawQuery
		cr.ServerToken = r.Header.Get("X-Postmark-Server-Token")
		cr.AccountToken = r.Header.Get("X-Postmark-Account-Token")
		cr.Body = body
		cr.Paths = append(cr.Paths, r.URL.Path)
		cr.mu.Unlock()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	defer srv.Close()

	t.Setenv("POSTMARK_API_BASE_URL", srv.URL)
	t.Setenv("POSTMARK_SERVER_TOKEN", "srv-tok")
	t.Setenv("POSTMARK_ACCOUNT_TOKEN", "acct-tok")
	t.Setenv("HOME", t.TempDir()) // isolate from the real ~/.config/pmx

	restore := silenceStdout()
	defer restore()

	resetFlags(rootCmd)
	rootCmd.SetArgs(args)
	return cr, rootCmd.Execute()
}
