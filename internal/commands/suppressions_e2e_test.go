package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func TestE2ESuppressionsList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Suppressions":[{"EmailAddress":"x@y.com"}]}`,
		"suppressions", "list")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/message-streams/outbound/suppressions/dump" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2ESuppressionsListStream(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Suppressions":[]}`,
		"suppressions", "list", "--stream", "broadcast")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/message-streams/broadcast/suppressions/dump" {
		t.Errorf("path = %s", cr.Path)
	}
}

func TestE2ESuppressionsCreateNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`,
		"suppressions", "create", "--email", "x@y.com")
	if err == nil {
		t.Fatal("expected gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2ESuppressionsCreate(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`,
		"suppressions", "create", "--email", "x@y.com", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/message-streams/outbound/suppressions" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(string(cr.Body), "EmailAddress") || !strings.Contains(string(cr.Body), "x@y.com") {
		t.Errorf("body missing EmailAddress: %s", cr.Body)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens wrong: srv=%q acct=%q", cr.ServerToken, cr.AccountToken)
	}
}

func TestE2ESuppressionsDeleteNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`,
		"suppressions", "delete", "--email", "x@y.com")
	if err == nil {
		t.Fatal("expected gate error without --yes")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2ESuppressionsDelete(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`,
		"suppressions", "delete", "--email", "x@y.com", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/message-streams/outbound/suppressions/delete" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(string(cr.Body), "EmailAddress") || !strings.Contains(string(cr.Body), "x@y.com") {
		t.Errorf("body missing EmailAddress: %s", cr.Body)
	}
	if cr.ServerToken != "srv-tok" || cr.AccountToken != "" {
		t.Errorf("scope tokens wrong: srv=%q acct=%q", cr.ServerToken, cr.AccountToken)
	}
}
