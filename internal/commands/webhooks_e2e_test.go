package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

// webhookBodyFile writes a valid JSON webhook payload to a temp file and returns
// its path, so create/edit get past readJSONFile and reach the confirm gate.
func webhookBodyFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "webhook.json")
	if err := os.WriteFile(p, []byte(`{"Url":"https://example.com/hook"}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}
	return p
}

func TestE2EWebhooksList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Webhooks":[{"ID":1}]}`, "webhooks", "list")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/webhooks" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EWebhooksListStream(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Webhooks":[{"ID":1}]}`, "webhooks", "list", "--stream", "outbound")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/webhooks" {
		t.Errorf("path = %s", cr.Path)
	}
	if !strings.Contains(cr.RawQuery, "MessageStream=outbound") {
		t.Errorf("query missing stream: %s", cr.RawQuery)
	}
}

func TestE2EWebhooksGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":123}`, "webhooks", "get", "123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/webhooks/123" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = %q %q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2EWebhooksCreateNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "webhooks", "create", "--file", webhookBodyFile(t))
	if err == nil {
		t.Fatal("expected gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EWebhooksCreateYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "webhooks", "create", "--file", webhookBodyFile(t), "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/webhooks" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(string(cr.Body), "example.com/hook") {
		t.Errorf("body not sent: %s", cr.Body)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = %q %q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2EWebhooksEditNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "webhooks", "edit", "123", "--file", webhookBodyFile(t))
	if err == nil {
		t.Fatal("expected gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EWebhooksEditYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "webhooks", "edit", "123", "--file", webhookBodyFile(t), "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/webhooks/123" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(string(cr.Body), "example.com/hook") {
		t.Errorf("body not sent: %s", cr.Body)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = %q %q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2EWebhooksDeleteNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "webhooks", "delete", "123")
	if err == nil {
		t.Fatal("expected gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EWebhooksDeleteYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "webhooks", "delete", "123", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "DELETE" || cr.Path != "/webhooks/123" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens = %q %q", cr.ServerToken, cr.AccountToken)
	}
}
