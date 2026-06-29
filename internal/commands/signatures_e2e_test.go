package commands

import (
	"errors"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func TestE2ESignaturesList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":1,"SenderSignatures":[{"ID":1}]}`,
		"signatures", "list")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/senders" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2ESignaturesGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":12345}`, "signatures", "get", "12345")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/senders/12345" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2ESignaturesCreateNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`, "signatures", "create", "--file", "-")
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

func TestE2ESignaturesCreateYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"signatures", "create", "--file", tmpJSON(t, `{"Name":"s","FromEmail":"a@example.com"}`), "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/senders" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2ESignaturesEditNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"signatures", "edit", "12345", "--file", "-")
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

func TestE2ESignaturesEditYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":1}`,
		"signatures", "edit", "12345", "--file", tmpJSON(t, `{"Name":"s"}`), "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PUT" || cr.Path != "/senders/12345" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2ESignaturesDeleteNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "signatures", "delete", "12345")
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

func TestE2ESignaturesDeleteYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "signatures", "delete", "12345", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "DELETE" || cr.Path != "/senders/12345" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2ESignaturesResendNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "signatures", "resend", "12345")
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

func TestE2ESignaturesResendYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "signatures", "resend", "12345", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/senders/12345/resend" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}

func TestE2ESignaturesRequestDKIMNoYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "signatures", "request-dkim", "12345")
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

func TestE2ESignaturesRequestDKIMYes(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "signatures", "request-dkim", "12345", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/senders/12345/requestnewdkim" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.AccountToken != "acct-tok" {
		t.Errorf("account token not sent: %q", cr.AccountToken)
	}
	if cr.ServerToken != "" {
		t.Errorf("server token must not be sent on account scope: %q", cr.ServerToken)
	}
}
