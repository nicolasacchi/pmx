package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

// writeBodyFile writes a JSON request body to a temp file and returns its path,
// used by the --file write commands so they don't read from (empty) stdin.
func writeBodyFile(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}
	return p
}

// domains_e2e_test.go — account-scope group. List/get are reads; create/edit/
// delete/verify-dkim/verify-returnpath/rotate-dkim are gated writes (a)+(b).

func assertDomainsAccountScope(t *testing.T, cr *capturedReq) {
	t.Helper()
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2EDomainsList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":1,"Domains":[{"ID":1}]}`, "domains", "list")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/domains" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}

func TestE2EDomainsGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":123456}`, "domains", "get", "123456")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/domains/123456" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}

func TestE2EDomainsCreateGated(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "domains", "create", "--file", "-")
	if err == nil {
		t.Fatal("expected confirm-gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("want APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EDomainsCreate(t *testing.T) {
	body := writeBodyFile(t, `{"Name":"example.com"}`)
	cr, err := runCmd(t, 200, `{"ID":1}`, "domains", "create", "--file", body, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/domains" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}

func TestE2EDomainsEditGated(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "domains", "edit", "123456", "--file", "-")
	if err == nil {
		t.Fatal("expected confirm-gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("want APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EDomainsEdit(t *testing.T) {
	body := writeBodyFile(t, `{"Name":"example.com"}`)
	cr, err := runCmd(t, 200, `{"ID":1}`, "domains", "edit", "123456", "--file", body, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/domains/123456" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}

func TestE2EDomainsDeleteGated(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "domains", "delete", "123456")
	if err == nil {
		t.Fatal("expected confirm-gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("want APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EDomainsDelete(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "domains", "delete", "123456", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "DELETE" || cr.Path != "/domains/123456" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}

func TestE2EDomainsVerifyDkimGated(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "domains", "verify-dkim", "123456")
	if err == nil {
		t.Fatal("expected confirm-gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("want APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EDomainsVerifyDkim(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "domains", "verify-dkim", "123456", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/domains/123456/verifyDkim" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}

func TestE2EDomainsVerifyReturnPathGated(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "domains", "verify-returnpath", "123456")
	if err == nil {
		t.Fatal("expected confirm-gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("want APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EDomainsVerifyReturnPath(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "domains", "verify-returnpath", "123456", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/domains/123456/verifyReturnPath" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}

func TestE2EDomainsRotateDkimGated(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "domains", "rotate-dkim", "123456")
	if err == nil {
		t.Fatal("expected confirm-gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("want APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EDomainsRotateDkim(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "domains", "rotate-dkim", "123456", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/domains/123456/rotatedkim" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	assertDomainsAccountScope(t, cr)
}
