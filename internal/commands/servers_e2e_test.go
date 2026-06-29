package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

// writeBody creates a temp JSON body file and returns its path, for the --file
// flag on gated write commands.
func writeBody(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatalf("write body: %v", err)
	}
	return p
}

func TestE2EServersList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":1,"Servers":[{"ID":1}]}`,
		"servers", "list", "--name", "production")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/servers" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(cr.RawQuery, "name=production") {
		t.Errorf("query missing name: %s", cr.RawQuery)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2EServersGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":12345}`, "servers", "get", "12345")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/servers/12345" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2EServersCreateNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "servers", "create", "--file", "-")
	if err == nil {
		t.Fatalf("expected gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError exit 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2EServersCreateYes(t *testing.T) {
	body := writeBody(t, `{"Name":"staging"}`)
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"servers", "create", "--file", body, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/servers" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2EServersEditNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"servers", "edit", "12345", "--file", "-")
	if err == nil {
		t.Fatalf("expected gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError exit 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2EServersEditYes(t *testing.T) {
	body := writeBody(t, `{"Color":"red"}`)
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"servers", "edit", "12345", "--file", body, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/servers/12345" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2EServersDeleteNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "servers", "delete", "12345")
	if err == nil {
		t.Fatalf("expected gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError exit 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2EServersDeleteYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "servers", "delete", "12345", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "DELETE" || cr.Path != "/servers/12345" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}
