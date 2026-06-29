package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func TestE2EInboundRulesList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":1,"InboundRules":[{"ID":1,"Rule":"x@y.com"}]}`,
		"inbound-rules", "list")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/triggers/inboundrules" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EInboundRulesCreateNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"inbound-rules", "create", "--rule", "x@y.com")
	if err == nil {
		t.Fatal("expected confirm-gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError with ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request is made")
	}
}

func TestE2EInboundRulesCreate(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"inbound-rules", "create", "--rule", "x@y.com", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/triggers/inboundrules" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(string(cr.Body), `"Rule"`) || !strings.Contains(string(cr.Body), "x@y.com") {
		t.Errorf("body missing Rule: %s", cr.Body)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EInboundRulesDeleteNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "inbound-rules", "delete", "12345")
	if err == nil {
		t.Fatal("expected confirm-gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError with ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Error("gate must block before any request is made")
	}
}

func TestE2EInboundRulesDelete(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "inbound-rules", "delete", "12345", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "DELETE" || cr.Path != "/triggers/inboundrules/12345" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}
