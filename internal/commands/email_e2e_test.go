package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func TestE2EEmailSendGateBlocks(t *testing.T) {
	cr, err := runCmd(t, 200, `{"MessageID":"a"}`,
		"email", "send", "--from", "no-reply@example.com", "--to", "me@example.com",
		"--subject", "smoke", "--text", "hello")
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

func TestE2EEmailSend(t *testing.T) {
	cr, err := runCmd(t, 200, `{"MessageID":"a"}`,
		"email", "send", "--from", "no-reply@example.com", "--to", "me@example.com",
		"--subject", "smoke", "--text", "hello", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/email" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	b := string(cr.Body)
	if !strings.Contains(b, `"From":"no-reply@example.com"`) {
		t.Errorf("body missing From: %s", b)
	}
	if !strings.Contains(b, `"To":"me@example.com"`) {
		t.Errorf("body missing To: %s", b)
	}
	if !strings.Contains(b, `"Subject":"smoke"`) {
		t.Errorf("body missing Subject: %s", b)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EEmailSendTemplateGateBlocks(t *testing.T) {
	cr, err := runCmd(t, 200, `{"MessageID":"a"}`,
		"email", "send-template", "--from", "no-reply@example.com", "--to", "me@example.com",
		"--alias", "welcome")
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

func TestE2EEmailSendTemplate(t *testing.T) {
	cr, err := runCmd(t, 200, `{"MessageID":"a"}`,
		"email", "send-template", "--from", "no-reply@example.com", "--to", "me@example.com",
		"--alias", "welcome", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/email/withTemplate" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	b := string(cr.Body)
	if !strings.Contains(b, `"From":"no-reply@example.com"`) {
		t.Errorf("body missing From: %s", b)
	}
	if !strings.Contains(b, `"To":"me@example.com"`) {
		t.Errorf("body missing To: %s", b)
	}
	if !strings.Contains(b, `"TemplateAlias":"welcome"`) {
		t.Errorf("body missing TemplateAlias: %s", b)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EEmailSendBatchGateBlocks(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "batch.json")
	if err := os.WriteFile(f, []byte(`[{"From":"no-reply@example.com","To":"me@example.com"}]`), 0o600); err != nil {
		t.Fatalf("write batch file: %v", err)
	}
	cr, err := runCmd(t, 200, `[{"MessageID":"a"}]`, "email", "send-batch", "--file", f)
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

func TestE2EEmailSendBatch(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "batch.json")
	if err := os.WriteFile(f, []byte(`[{"From":"no-reply@example.com","To":"me@example.com"}]`), 0o600); err != nil {
		t.Fatalf("write batch file: %v", err)
	}
	cr, err := runCmd(t, 200, `[{"MessageID":"a"}]`, "email", "send-batch", "--file", f, "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/email/batch" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	b := string(cr.Body)
	if !strings.Contains(b, `"From":"no-reply@example.com"`) {
		t.Errorf("body missing From: %s", b)
	}
	if !strings.Contains(b, `"To":"me@example.com"`) {
		t.Errorf("body missing To: %s", b)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}
