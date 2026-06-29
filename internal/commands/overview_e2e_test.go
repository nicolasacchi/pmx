package commands

import (
	"testing"
)

func TestE2EOverview(t *testing.T) {
	cr, err := runCmd(t, 200, `{}`, "overview")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !cr.hitPath("/deliverystats") {
		t.Errorf("missing /deliverystats, paths = %v", cr.Paths)
	}
	if !cr.hitPath("/stats/outbound") {
		t.Errorf("missing /stats/outbound, paths = %v", cr.Paths)
	}
	if !cr.hitPath("/bounces") {
		t.Errorf("missing /bounces, paths = %v", cr.Paths)
	}
	if !cr.hitPath("/message-streams") {
		t.Errorf("missing /message-streams, paths = %v", cr.Paths)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
}
