package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func TestE2EServerGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Name":"my-server","TrackOpens":true}`, "server", "get")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/server" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EServerEditGateBlocks(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(f, []byte(`{"TrackOpens":true}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}
	cr, err := runCmd(t, 200, `{"Name":"my-server"}`, "server", "edit", "--file", f)
	if err == nil {
		t.Fatal("expected gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("want APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EServerEdit(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(f, []byte(`{"TrackOpens":true}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}
	cr, err := runCmd(t, 200, `{"Name":"my-server","TrackOpens":true}`,
		"server", "edit", "--file", f, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PATCH" || cr.Path != "/server" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(string(cr.Body), `"TrackOpens":true`) {
		t.Errorf("body missing TrackOpens: %s", cr.Body)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}
