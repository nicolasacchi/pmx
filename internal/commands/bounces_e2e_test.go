package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func TestE2EBouncesList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":1,"Bounces":[{"ID":1}]}`,
		"bounces", "list", "--inactive", "--type", "HardBounce",
		"--email", "x@y.com", "--tag", "otp")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/bounces" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(cr.RawQuery, "type=HardBounce") {
		t.Errorf("query missing type: %s", cr.RawQuery)
	}
	if !strings.Contains(cr.RawQuery, "emailFilter=x%40y.com") {
		t.Errorf("query missing emailFilter: %s", cr.RawQuery)
	}
	if !strings.Contains(cr.RawQuery, "tag=otp") {
		t.Errorf("query missing tag: %s", cr.RawQuery)
	}
	if !strings.Contains(cr.RawQuery, "inactive=true") {
		t.Errorf("query missing inactive: %s", cr.RawQuery)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EBouncesGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":123}`, "bounces", "get", "123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/bounces/123" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope wrong: srv=%q acct=%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2EBouncesDump(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Body":"raw"}`, "bounces", "dump", "123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/bounces/123/dump" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
}

func TestE2EBouncesActivateNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Bounce":{"ID":123}}`, "bounces", "activate", "123")
	if err == nil {
		t.Fatal("expected error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *client.APIError, got %T: %v", err, err)
	}
	if apiErr.ExitCode() != 6 {
		t.Errorf("exit code = %d, want 6", apiErr.ExitCode())
	}
	if cr.Hit {
		t.Error("gate must block before any request")
	}
}

func TestE2EBouncesActivateConfirmed(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Bounce":{"ID":123}}`, "bounces", "activate", "123", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/bounces/123/activate" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope wrong: srv=%q acct=%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2EBouncesTags(t *testing.T) {
	cr, err := runCmd(t, 200, `["otp","welcome"]`, "bounces", "tags")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/bounces/tags" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope wrong: srv=%q acct=%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2EBouncesDeliveryStats(t *testing.T) {
	cr, err := runCmd(t, 200, `{"InactiveMails":2,"Bounces":[]}`, "bounces", "deliverystats")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/deliverystats" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope wrong: srv=%q acct=%q", cr.ServerToken, cr.AccountToken)
	}
}
