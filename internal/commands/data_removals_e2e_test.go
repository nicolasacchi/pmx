package commands

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

const dataRemovalBody = `{"RequestedBy":"ops@example.com","RequestedByEmail":"ops@example.com","RecipientEmailAddress":"customer@example.com","NotifyWhenCompleted":true}`

func TestE2EDataRemovalsCreateGateBlocks(t *testing.T) {
	file := filepath.Join(t.TempDir(), "removal.json")
	if err := os.WriteFile(file, []byte(dataRemovalBody), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cr, err := runCmd(t, 200, `{"ID":42}`, "data-removals", "create", "--file", file)
	if err == nil {
		t.Fatal("expected gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *client.APIError: %v", err)
	}
	if apiErr.ExitCode() != 6 {
		t.Errorf("exit code = %d, want 6", apiErr.ExitCode())
	}
	if cr.Hit {
		t.Error("gate should block before any request")
	}
}

func TestE2EDataRemovalsCreate(t *testing.T) {
	file := filepath.Join(t.TempDir(), "removal.json")
	if err := os.WriteFile(file, []byte(dataRemovalBody), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cr, err := runCmd(t, 200, `{"ID":42}`, "data-removals", "create", "--file", file, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/data-removals" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	// The CLI re-marshals the JSON (key order is not preserved), so compare the
	// decoded body rather than raw bytes.
	var got, want map[string]any
	_ = json.Unmarshal(cr.Body, &got)
	_ = json.Unmarshal([]byte(dataRemovalBody), &want)
	if got["RecipientEmailAddress"] != want["RecipientEmailAddress"] || got["RequestedBy"] != want["RequestedBy"] {
		t.Errorf("body round-trip mismatch: %s", cr.Body)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2EDataRemovalsStatus(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":42,"Status":"Pending"}`, "data-removals", "status", "42")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/data-removals/42" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}
