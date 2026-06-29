package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func writeTempTemplate(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tpl.json")
	if err := os.WriteFile(p, []byte(`{"Name":"Welcome","Alias":"welcome","Subject":"Hi","HtmlBody":"<p>Hi</p>"}`), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

func TestE2ETemplatesList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":1,"Templates":[{"TemplateId":1}]}`, "templates", "list")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/templates" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2ETemplatesGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TemplateId":1234567}`, "templates", "get", "1234567")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/templates/1234567" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = srv:%q acct:%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2ETemplatesCreateGated(t *testing.T) {
	file := writeTempTemplate(t)
	cr, err := runCmd(t, 200, `{"TemplateId":1}`, "templates", "create", "--file", file)
	if err == nil {
		t.Fatalf("expected gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError exit code 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate should block before any request")
	}
}

func TestE2ETemplatesCreate(t *testing.T) {
	file := writeTempTemplate(t)
	cr, err := runCmd(t, 200, `{"TemplateId":1}`, "templates", "create", "--file", file, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/templates" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if len(cr.Body) == 0 {
		t.Errorf("expected request body from --file")
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = srv:%q acct:%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2ETemplatesEditGated(t *testing.T) {
	file := writeTempTemplate(t)
	cr, err := runCmd(t, 200, `{"TemplateId":1}`, "templates", "edit", "1234567", "--file", file)
	if err == nil {
		t.Fatalf("expected gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError exit code 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate should block before any request")
	}
}

func TestE2ETemplatesEdit(t *testing.T) {
	file := writeTempTemplate(t)
	cr, err := runCmd(t, 200, `{"TemplateId":1}`, "templates", "edit", "1234567", "--file", file, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/templates/1234567" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if len(cr.Body) == 0 {
		t.Errorf("expected request body from --file")
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = srv:%q acct:%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2ETemplatesDeleteGated(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "templates", "delete", "1234567")
	if err == nil {
		t.Fatalf("expected gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError exit code 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate should block before any request")
	}
}

func TestE2ETemplatesDelete(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "templates", "delete", "1234567", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "DELETE" || cr.Path != "/templates/1234567" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = srv:%q acct:%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2ETemplatesValidate(t *testing.T) {
	file := writeTempTemplate(t)
	cr, err := runCmd(t, 200, `{"AllContentIsValid":true}`, "templates", "validate", "--file", file)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/templates/validate" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if len(cr.Body) == 0 {
		t.Errorf("expected request body from --file")
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = srv:%q acct:%q", cr.ServerToken, cr.AccountToken)
	}
}
