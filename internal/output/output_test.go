package output

import (
	"encoding/json"
	"testing"
)

func TestIsJSON(t *testing.T) {
	// --json or --jq always forces JSON regardless of TTY.
	if !IsJSON(true, "") {
		t.Error("--json should force JSON")
	}
	if !IsJSON(false, "#.MessageID") {
		t.Error("--jq should force JSON")
	}
	// In a test process stdout is not a TTY, so the default is also JSON.
	if !IsJSON(false, "") {
		t.Error("non-TTY should default to JSON")
	}
}

func TestApplyFilterProjection(t *testing.T) {
	data := json.RawMessage(`[{"MessageID":"a","Status":"Sent"},{"MessageID":"b","Status":"Bounced"}]`)
	got, err := ApplyFilter(data, "#.MessageID")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(got) != `["a","b"]` {
		t.Errorf("got %s, want [\"a\",\"b\"]", got)
	}
}

func TestApplyFilterMissingPath(t *testing.T) {
	got, err := ApplyFilter(json.RawMessage(`{"a":1}`), "nope")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(got) != "null" {
		t.Errorf("missing path should yield null, got %s", got)
	}
}

func TestApplyFilterEmptyPassthrough(t *testing.T) {
	in := json.RawMessage(`{"a":1}`)
	got, err := ApplyFilter(in, "")
	if err != nil || string(got) != string(in) {
		t.Errorf("empty filter should pass through unchanged: %s %v", got, err)
	}
}

func TestFormatValue(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, ""},
		{"hello", "hello"},
		{float64(42), "42"},
		{float64(3.5), "3.50"},
		{true, "true"},
		{false, "false"},
		{[]any{"a", "b"}, "a, b"},
	}
	for _, c := range cases {
		if got := formatValue(c.in); got != c.want {
			t.Errorf("formatValue(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTruncate60(t *testing.T) {
	long := ""
	for i := 0; i < 80; i++ {
		long += "x"
	}
	got := truncate60(long)
	if len(got) != 60 {
		t.Errorf("truncate60 len = %d, want 60", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Errorf("truncated value should end with ...: %q", got)
	}
}

func TestPrintTableKnownCommand(t *testing.T) {
	// A known command with PascalCase Postmark keys should render without error.
	data := json.RawMessage(`[{"MessageID":"abc","Recipients":["x@y.com"],"Subject":"Code","Status":"Sent","ReceivedAt":"2026-06-29"}]`)
	if err := printTable("messages.outbound", data); err != nil {
		t.Errorf("printTable(messages.outbound) error: %v", err)
	}
}

func TestPrintTableUnknownCommandErrors(t *testing.T) {
	// An unknown command must error so PrintData falls back to JSON.
	if err := printTable("does.not.exist", json.RawMessage(`[]`)); err == nil {
		t.Error("expected error for unknown table command")
	}
}
