package commands

import (
	"strings"
	"testing"
)

// stats group (SERVER scope) — every subcommand GETs an aggregate object under
// /stats/outbound[/<sub>]. Assert path + scope token, and that --tag adds the
// tag=... query param.

func TestE2EStatsOutbound(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Sent":10,"Bounced":1}`, "stats", "outbound")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Method != "GET" || cr.Path != "/stats/outbound" {
		t.Errorf("method/path = %s %s", cr.Method, cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}

func TestE2EStatsSends(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Sent":42}`, "stats", "sends")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/stats/outbound/sends" {
		t.Errorf("path = %s", cr.Path)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
}

func TestE2EStatsBounces(t *testing.T) {
	cr, err := runCmd(t, 200, `{"HardBounce":1}`, "stats", "bounces")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/stats/outbound/bounces" {
		t.Errorf("path = %s", cr.Path)
	}
}

func TestE2EStatsSpam(t *testing.T) {
	cr, err := runCmd(t, 200, `{"SpamComplaint":0}`, "stats", "spam")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/stats/outbound/spam" {
		t.Errorf("path = %s", cr.Path)
	}
}

func TestE2EStatsOpens(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Opens":5}`, "stats", "opens")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/stats/outbound/opens" {
		t.Errorf("path = %s", cr.Path)
	}
}

func TestE2EStatsClicks(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Clicks":3}`, "stats", "clicks")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/stats/outbound/clicks" {
		t.Errorf("path = %s", cr.Path)
	}
}

func TestE2EStatsTagFilter(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Opens":5}`, "stats", "opens", "--tag", "otp")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cr.Path != "/stats/outbound/opens" {
		t.Errorf("path = %s", cr.Path)
	}
	if !strings.Contains(cr.RawQuery, "tag=otp") {
		t.Errorf("query missing tag: %s", cr.RawQuery)
	}
}
