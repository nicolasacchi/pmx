package commands

import (
	"errors"
	"testing"

	"github.com/nicolasacchi/pmx/internal/client"
)

func TestE2EStreamsList(t *testing.T) {
	cr, err := runCmd(t, 200, `{"MessageStreams":[{"ID":"outbound"}]}`, "streams", "list")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/message-streams" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EStreamsGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":"broadcast"}`, "streams", "get", "broadcast")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/message-streams/broadcast" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
}

func TestE2EStreamsCreateNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "streams", "create", "--file", "-")
	if err == nil {
		t.Fatal("expected confirm-gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2EStreamsCreateConfirmed(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":"announcements"}`,
		"streams", "create", "--file", tmpJSON(t, `{"ID":"announcements","Name":"A","MessageStreamType":"Broadcasts"}`), "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/message-streams" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EStreamsEditNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "streams", "edit", "broadcast", "--file", "-")
	if err == nil {
		t.Fatal("expected confirm-gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2EStreamsEditConfirmed(t *testing.T) {
	cr, err := runCmd(t, 200, `{"ID":"broadcast"}`,
		"streams", "edit", "broadcast", "--file", tmpJSON(t, `{"Name":"B"}`), "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "PATCH" || cr.Path != "/message-streams/broadcast" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EStreamsArchiveNoConfirm(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "streams", "archive", "old-broadcasts")
	if err == nil {
		t.Fatal("expected confirm-gate error, got nil")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
		t.Errorf("expected APIError ExitCode 6, got %v", err)
	}
	if cr.Hit {
		t.Errorf("gate must block before any request")
	}
}

func TestE2EStreamsArchiveConfirmed(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "streams", "archive", "old-broadcasts", "--yes")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "POST" || cr.Path != "/message-streams/old-broadcasts/archive" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
}
