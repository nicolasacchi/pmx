package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/nicolasacchi/pmx/internal/client"
)

// extractList unwraps a Postmark list envelope {"TotalCount":N,"<key>":[...]}
// into the bare array, noting the total on stderr. This keeps --jq paths simple
// and lets the agent row-cap (which only caps top-level arrays) apply.
func extractList(data json.RawMessage, key string) json.RawMessage {
	var wrapper map[string]json.RawMessage
	if json.Unmarshal(data, &wrapper) != nil {
		return data
	}
	arr, ok := wrapper[key]
	if !ok {
		return data
	}
	if tc, ok := wrapper["TotalCount"]; ok && !quietFlag {
		var n int
		if json.Unmarshal(tc, &n) == nil && n > 0 {
			fmt.Fprintf(os.Stderr, "%d total\n", n)
		}
	}
	return arr
}

// jsonLen returns the element count of a JSON array (0 if not an array).
func jsonLen(data json.RawMessage) int {
	var arr []json.RawMessage
	if json.Unmarshal(data, &arr) == nil {
		return len(arr)
	}
	return 0
}

// pageParams seeds count/offset from the global --limit/--offset flags. Postmark
// caps count at 500/page; we pass the user's limit straight through.
func pageParams() url.Values {
	p := url.Values{}
	count := limitFlag
	if count <= 0 {
		count = 100
	}
	if count > 500 {
		count = 500
	}
	p.Set("count", strconv.Itoa(count))
	p.Set("offset", strconv.Itoa(offsetFlag))
	return p
}

// addDateParams sets Postmark's fromdate/todate from --from/--to when present.
func addDateParams(p url.Values) error {
	from, err := fromDate()
	if err != nil {
		return err
	}
	to, err := toDate()
	if err != nil {
		return err
	}
	if from != "" {
		p.Set("fromdate", from)
	}
	if to != "" {
		p.Set("todate", to)
	}
	return nil
}

// gateWrite is the standard preamble for a destructive verb: --dry-run prints
// and returns handled=true; otherwise the confirm gate runs. Returns handled to
// tell the caller to return immediately (with err).
func gateWrite(action string) (bool, error) {
	if dryRun() {
		fmt.Fprintf(os.Stdout, "--dry-run: would %s, no changes made\n", action)
		return true, nil
	}
	if err := requireConfirm(action); err != nil {
		return true, err
	}
	return false, nil
}

// ctx is the shared request context (no global deadline; the client sets its own
// per-request timeout).
func ctx() context.Context { return context.Background() }

// readJSONFile decodes a JSON request body from a file path ("-" reads stdin),
// used by the `--file` write commands. The decoded value is returned as `any`
// so it round-trips through the client's json.Marshal unchanged.
func readJSONFile(path string) (any, error) {
	var raw []byte
	var err error
	if path == "-" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("read body file: %w", err)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("parse JSON body from %s: %w", path, err)
	}
	return v, nil
}

// serverGet / accountGet are thin helpers so command files read cleanly.
func serverGet(c *client.Client, path string, params url.Values) (json.RawMessage, error) {
	return c.Get(ctx(), client.ScopeServer, path, params)
}

func accountGet(c *client.Client, path string, params url.Values) (json.RawMessage, error) {
	return c.Get(ctx(), client.ScopeAccount, path, params)
}
