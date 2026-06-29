package commands

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractList(t *testing.T) {
	quietFlag = true
	defer func() { quietFlag = false }()

	data := json.RawMessage(`{"TotalCount":2,"Messages":[{"MessageID":"a"},{"MessageID":"b"}]}`)
	got := extractList(data, "Messages")
	if jsonLen(got) != 2 {
		t.Errorf("extractList len = %d, want 2", jsonLen(got))
	}

	// Missing key → original data unchanged.
	same := extractList(json.RawMessage(`{"X":1}`), "Messages")
	if string(same) != `{"X":1}` {
		t.Errorf("missing key should pass through, got %s", same)
	}
}

func TestJSONLen(t *testing.T) {
	if jsonLen(json.RawMessage(`[1,2,3]`)) != 3 {
		t.Error("array len")
	}
	if jsonLen(json.RawMessage(`{"a":1}`)) != 0 {
		t.Error("non-array should be 0")
	}
}

func TestPageParams(t *testing.T) {
	defer func() { limitFlag, offsetFlag = 50, 0 }()

	limitFlag, offsetFlag = 25, 100
	p := pageParams()
	if p.Get("count") != "25" || p.Get("offset") != "100" {
		t.Errorf("pageParams = %v", p)
	}

	limitFlag = 1000 // over Postmark's 500 cap
	if pageParams().Get("count") != "500" {
		t.Errorf("count should cap at 500, got %s", pageParams().Get("count"))
	}

	limitFlag = 0 // unset → sane default
	if pageParams().Get("count") != "100" {
		t.Errorf("count default = %s, want 100", pageParams().Get("count"))
	}
}

func TestAddDateParams(t *testing.T) {
	defer func() { fromFlag, toFlag = "", "" }()

	fromFlag, toFlag = "2026-06-01", "2026-06-29"
	p := url.Values{}
	if err := addDateParams(p); err != nil {
		t.Fatalf("err: %v", err)
	}
	if p.Get("fromdate") != "2026-06-01" || p.Get("todate") != "2026-06-29" {
		t.Errorf("date params = %v", p)
	}

	// Unset flags add nothing.
	fromFlag, toFlag = "", ""
	p2 := url.Values{}
	if err := addDateParams(p2); err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(p2) != 0 {
		t.Errorf("no date flags should add no params, got %v", p2)
	}

	// Invalid value surfaces an error.
	fromFlag = "garbage"
	if err := addDateParams(url.Values{}); err == nil {
		t.Error("invalid --from should error")
	}
}

func TestSetIfNonEmpty(t *testing.T) {
	p := url.Values{}
	setIfNonEmpty(p, "a", "x")
	setIfNonEmpty(p, "b", "")
	if p.Get("a") != "x" {
		t.Error("non-empty value should be set")
	}
	if _, ok := p["b"]; ok {
		t.Error("empty value should not be set")
	}
}

func TestFirstMessageID(t *testing.T) {
	if got := firstMessageID(json.RawMessage(`[{"MessageID":"a"},{"MessageID":"b"}]`)); got != "a" {
		t.Errorf("got %q, want a", got)
	}
	if got := firstMessageID(json.RawMessage(`[]`)); got != "" {
		t.Errorf("empty array should yield empty id, got %q", got)
	}
}

func TestReadJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	if err := os.WriteFile(path, []byte(`{"From":"a@b.com"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	v, err := readJSONFile(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	m, ok := v.(map[string]any)
	if !ok || m["From"] != "a@b.com" {
		t.Errorf("parsed body = %v", v)
	}

	if _, err := readJSONFile(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("missing file should error")
	}

	bad := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(bad, []byte(`{not json`), 0o600)
	if _, err := readJSONFile(bad); err == nil {
		t.Error("invalid JSON should error")
	}
}
