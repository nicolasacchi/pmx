package commands

import (
	"strings"
	"testing"
)

// EXEMPLAR E2E test — every group's *_e2e_test.go mirrors this shape: drive the
// real command via runCmd, then assert on the captured request (method, path,
// query, scope token, body) and on the exit code for gated writes.

func TestE2EMessagesOutbound(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":1,"Messages":[{"MessageID":"a"}]}`,
		"messages", "outbound", "--recipient", "x@y.com", "--status", "bounced")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/messages/outbound" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if !strings.Contains(cr.RawQuery, "recipient=x%40y.com") {
		t.Errorf("query missing recipient: %s", cr.RawQuery)
	}
	if !strings.Contains(cr.RawQuery, "status=bounced") {
		t.Errorf("query missing status: %s", cr.RawQuery)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EMessagesOutboundGet(t *testing.T) {
	cr, err := runCmd(t, 200, `{"MessageID":"abc"}`, "messages", "outbound", "get", "abc123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/messages/outbound/abc123/details" {
		t.Errorf("path = %s", cr.Path)
	}
}

func TestE2EMessagesEventsHitsDetails(t *testing.T) {
	cr, err := runCmd(t, 200, `{"MessageID":"a","MessageEvents":[{"Type":"Delivered"}]}`,
		"messages", "outbound", "events", "abc123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/messages/outbound/abc123/details" {
		t.Errorf("events should read details, path = %s", cr.Path)
	}
}

func TestE2EMessagesInbound(t *testing.T) {
	cr, err := runCmd(t, 200, `{"TotalCount":0,"InboundMessages":[]}`, "messages", "inbound")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/messages/inbound" {
		t.Errorf("path = %s", cr.Path)
	}
}
